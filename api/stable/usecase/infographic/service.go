package infographic

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"webdesa/api/config"
	"webdesa/api/domain/infographic"
	"webdesa/api/pkg/clock"
	"webdesa/api/pkg/pagination"

	"github.com/google/uuid"
)

// PublicTokenTTL is the expiry window for tokens generated for the
// public consumption path. Shorter than the admin preview token (10m)
// to limit the blast radius of a leaked JWT.
const PublicTokenTTL = 5 * time.Minute

// PerComponentRateLimit is the maximum number of token generations allowed
// from a single IP for a single infographic within the rate window.
const PerComponentRateLimit = 30

// PerComponentRateWindow is the time window for the per-component rate limit.
const PerComponentRateWindow = time.Hour

// Service implements infographic management business logic.
// It accepts interfaces (Repository, CategoryLookup, AccessLogRepository) and
// returns concrete structs (Infographic). This follows the "accept interfaces,
// return structs" Go idiom.
type Service struct {
	repo           Repository
	categoryRepo   CategoryLookup
	accessLogRepo  AccessLogRepository
	clock          clock.Clock
	metabaseConfig config.MetabaseConfig
}

// NewService creates a new infographic management service.
// Dependencies are injected via constructor following Clean Architecture principles.
func NewService(repo Repository, clk clock.Clock, metabaseConfig config.MetabaseConfig, categoryRepo CategoryLookup, accessLogRepo AccessLogRepository) *Service {
	return &Service{
		repo:           repo,
		categoryRepo:   categoryRepo,
		accessLogRepo:  accessLogRepo,
		clock:          clk,
		metabaseConfig: metabaseConfig,
	}
}

// CreateInfographicInput represents the input for creating an infographic
type CreateInfographicInput struct {
	ComponentID     int64
	ComponentType   string
	SectionName     string
	SectionEndpoint string
	Category        *string // Optional UUID of an existing infographic category
	State           bool
}

// UpdateInfographicInput represents the input for updating an infographic
type UpdateInfographicInput struct {
	ComponentID     int64
	ComponentType   string
	SectionName     string
	SectionEndpoint string
	Category        *string // Optional UUID; pass nil to leave unchanged
	State           bool
}

// ListInfographicsInput represents the input for listing infographics
type ListInfographicsInput struct {
	SectionName *string
	State       *bool
	Query       *string
	Category    *string
	Page        int
	Limit       int
}

// Create creates a new infographic.
// Returns concrete Infographic struct.
func (s *Service) Create(ctx context.Context, input CreateInfographicInput) (*infographic.Infographic, error) {
	// Validate the category exists, if one is provided. The FK is enforced
	// at the DB level, but checking here provides a friendlier 400 response.
	if input.Category != nil && *input.Category != "" {
		if _, err := s.categoryRepo.FindByID(ctx, *input.Category); err != nil {
			return nil, fmt.Errorf("invalid category: %w", err)
		}
	}

	now := s.clock.Now()
	i := &infographic.Infographic{
		ID:              uuid.New().String(),
		ComponentID:     input.ComponentID,
		ComponentType:   infographic.ComponentType(input.ComponentType),
		SectionName:     input.SectionName,
		SectionEndpoint: input.SectionEndpoint,
		Category:        input.Category,
		State:           input.State,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Validate domain invariants
	if err := i.Validate(); err != nil {
		return nil, fmt.Errorf("invalid infographic data: %w", err)
	}

	// Persist to database
	if err := s.repo.Create(ctx, i); err != nil {
		return nil, fmt.Errorf("failed to create infographic: %w", err)
	}

	return i, nil
}

// GetByID retrieves an infographic by ID.
// Returns error if infographic not found.
func (s *Service) GetByID(ctx context.Context, id string) (*infographic.Infographic, error) {
	i, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get infographic: %w", err)
	}
	return i, nil
}

// LogAccess records a token-issuance audit log entry. Best-effort: errors
// are returned to the caller but not propagated to the response.
func (s *Service) LogAccess(ctx context.Context, infographicID string, componentID int64, componentType, endpoint string, meta RequestMeta, ttl time.Duration) error {
	if s.accessLogRepo == nil {
		return nil
	}
	infID, err := uuid.Parse(infographicID)
	if err != nil {
		return err
	}
	now := s.clock.Now()
	ip := parseIP(meta.IPAddress)
	return s.accessLogRepo.Create(ctx, &infographic.AccessLog{
		InfographicID:  infID,
		ComponentID:    componentID,
		ComponentType:  componentType,
		Endpoint:       endpoint,
		IPAddress:      ip,
		UserAgent:      meta.UserAgent,
		Referer:        meta.Referer,
		TokenIssuedAt:  now,
		TokenExpiresAt: now.Add(ttl),
		CreatedAt:      now,
	})
}

// RequestMeta carries the caller's IP, User-Agent, and Referer for audit logging
// and rate limiting. Handlers extract these from the http.Request.
type RequestMeta struct {
	IPAddress string
	UserAgent string
	Referer   string
}

// GetPublicAccess retrieves the infographic, checks the per-component rate
// limit for the caller, generates a short-lived Metabase JWT, and writes an
// access-log entry. Returns the infographic, the token, and its expiry time.
func (s *Service) GetPublicAccess(ctx context.Context, id string, meta RequestMeta) (*infographic.Infographic, string, time.Time, error) {
	i, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, "", time.Time{}, fmt.Errorf("failed to get infographic: %w", err)
	}

	// Per-component rate limit (per IP per infographic within 1h).
	if s.accessLogRepo != nil && meta.IPAddress != "" {
		infID, parseErr := uuid.Parse(id)
		if parseErr == nil {
			since := s.clock.Now().Add(-PerComponentRateWindow)
			n, countErr := s.accessLogRepo.CountByInfographicAndIPSince(ctx, infID, meta.IPAddress, since)
			if countErr == nil && n >= PerComponentRateLimit {
				return nil, "", time.Time{}, fmt.Errorf("rate limit exceeded: max %d token generations per hour for this infographic", PerComponentRateLimit)
			}
		}
	}

	// Generate short-lived token for public consumption.
	expiresAt := s.clock.Now().Add(PublicTokenTTL)
	token, err := s.GenerateMetabaseTokenWithExpiry(i.ComponentID, string(i.ComponentType), expiresAt)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	// Audit log entry (best-effort, does not block the response).
	if s.accessLogRepo != nil {
		infID, parseErr := uuid.Parse(id)
		if parseErr == nil {
			now := s.clock.Now()
			ipStr := parseIP(meta.IPAddress)
			logEntry := &infographic.AccessLog{
				InfographicID:  infID,
				ComponentID:    i.ComponentID,
				ComponentType:  string(i.ComponentType),
				Endpoint:       "public_detail",
				IPAddress:      ipStr,
				UserAgent:      meta.UserAgent,
				Referer:        meta.Referer,
				TokenIssuedAt:  now,
				TokenExpiresAt: expiresAt,
				CreatedAt:      now,
			}
			// Sanitize strings — strip NUL bytes and invalid UTF-8 sequences
			logEntry.UserAgent = sanitizeString(meta.UserAgent)
			logEntry.Referer = sanitizeString(meta.Referer)
			_ = s.accessLogRepo.Create(ctx, logEntry)
		}
	}

	return i, token, expiresAt, nil
}

// ListAccessLogs returns paginated access logs filtered by infographic or IP.
func (s *Service) ListAccessLogs(ctx context.Context, infographicID *uuid.UUID, ip string, page, limit int) ([]*infographic.AccessLog, int, error) {
	if s.accessLogRepo == nil {
		return nil, 0, nil
	}
	if infographicID != nil {
		offset, validatedLimit, err := pagination.Paginate(page, limit)
		if err != nil {
			return nil, 0, err
		}
		return s.accessLogRepo.ListByInfographic(ctx, *infographicID, offset, validatedLimit)
	}
	if ip != "" {
		offset, validatedLimit, err := pagination.Paginate(page, limit)
		if err != nil {
			return nil, 0, err
		}
		return s.accessLogRepo.ListByIP(ctx, ip, offset, validatedLimit)
	}
	return nil, 0, fmt.Errorf("either infographic_id or ip must be provided")
}

func parseIP(s string) string {
	if s == "" {
		return ""
	}
	ip := net.ParseIP(s)
	if ip == nil {
		return ""
	}
	return ip.String()
}

// sanitizeString strips NUL bytes and other invalid UTF-8 sequences that
// would cause "pq: invalid byte sequence" errors on PostgreSQL TEXT columns.
func sanitizeString(s string) string {
	if s == "" {
		return s
	}
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r == 0 || r == 0xFFFD {
			continue
		}
		out = append(out, r)
	}
	return string(out)
}

// List retrieves paginated infographics with optional filtering.
// Returns infographics, pagination info, and error.
func (s *Service) List(ctx context.Context, input ListInfographicsInput) ([]*infographic.Infographic, pagination.Result, error) {
	// Validate and normalize pagination parameters
	offset, validatedLimit, err := pagination.Paginate(input.Page, input.Limit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("invalid pagination parameters: %w", err)
	}

	// Query repository
	infographics, total, err := s.repo.List(ctx, input.SectionName, input.State, input.Query, input.Category, offset, validatedLimit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("failed to list infographics: %w", err)
	}

	// Create pagination metadata
	paginationResult := pagination.NewResult(input.Page, validatedLimit, total)

	return infographics, paginationResult, nil
}

// Update updates an existing infographic.
// Returns updated Infographic struct.
func (s *Service) Update(ctx context.Context, id string, input UpdateInfographicInput) (*infographic.Infographic, error) {
	// Validate the new category exists, if one is provided.
	if input.Category != nil && *input.Category != "" {
		if _, err := s.categoryRepo.FindByID(ctx, *input.Category); err != nil {
			return nil, fmt.Errorf("invalid category: %w", err)
		}
	}

	// Get existing infographic
	i, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get infographic: %w", err)
	}

	// Update fields
	i.ComponentID = input.ComponentID
	i.ComponentType = infographic.ComponentType(input.ComponentType)
	i.SectionName = input.SectionName
	i.SectionEndpoint = input.SectionEndpoint
	if input.Category != nil {
		i.Category = input.Category
	}
	i.State = input.State
	i.UpdatedAt = s.clock.Now()

	// Validate domain invariants
	if err := i.Validate(); err != nil {
		return nil, fmt.Errorf("invalid infographic data: %w", err)
	}

	// Persist to database
	if err := s.repo.Update(ctx, i); err != nil {
		return nil, fmt.Errorf("failed to update infographic: %w", err)
	}

	return i, nil
}

// Delete removes an infographic.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete infographic: %w", err)
	}
	return nil
}

// GetSectionNames retrieves all unique section names.
func (s *Service) GetSectionNames(ctx context.Context) ([]string, error) {
	names, err := s.repo.GetSectionNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get section names: %w", err)
	}
	return names, nil
}

// GenerateMetabaseToken generates a JWT token for Metabase authentication.
// It creates a token for accessing a specific component (question or dashboard).
// ComponentID must be an integer as required by Metabase.
// Token TTL is 10 minutes for admin previews.
func (s *Service) GenerateMetabaseToken(componentID int64, componentType string) (string, error) {
	return s.GenerateMetabaseTokenWithExpiry(componentID, componentType, s.clock.Now().Add(10*time.Minute))
}

// GenerateMetabaseTokenWithExpiry signs a Metabase JWT that expires at the
// supplied time. Use PublicTokenTTL for public consumption, 10m for admin.
func (s *Service) GenerateMetabaseTokenWithExpiry(componentID int64, componentType string, expiresAt time.Time) (string, error) {
	// Determine resource type based on component_type
	resourceType := "question"
	if componentType == "dashboard" {
		resourceType = "dashboard"
	}

	// Create JWT payload with integer component ID
	payload := map[string]interface{}{
		"resource": map[string]interface{}{
			resourceType: componentID,
		},
		"params": map[string]interface{}{},
		"exp":    expiresAt.Unix(),
	}

	// Encode payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create header
	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}

	// Encode header and payload to base64
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create signature
	message := headerB64 + "." + payloadB64
	h256 := hmac.New(sha256.New, []byte(s.metabaseConfig.SecretKey))
	h256.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h256.Sum(nil))

	// Return complete JWT token
	return message + "." + signature, nil
}
