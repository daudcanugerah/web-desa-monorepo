package ppid

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"webdesa/api/domain/ppid"
	"webdesa/api/pkg/clock"
	"webdesa/api/pkg/pagination"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Service implements PPID (public information disclosure) management business logic.
// It accepts interfaces (Repository, FileHandler, EmailService) and returns concrete structs (PPID, PPIDRequest).
// This follows the "accept interfaces, return structs" Go idiom.
type Service struct {
	repo         Repository
	fileHandler  FileHandler
	emailService EmailService
	categoryRepo CategoryLookup
	clock        clock.Clock
	jwtSecret    string
	emailConfig  EmailConfig
}

// EmailConfig holds email template customization from database
type EmailConfig struct {
	VillageName  string
	SupportEmail string
	WebsiteURL   string
	DomainAddr   string
}

// NewService creates a new PPID management service.
// Dependencies are injected via constructor following Clean Architecture principles.
func NewService(repo Repository, fileHandler FileHandler, clk clock.Clock, jwtSecret string, categoryRepo CategoryLookup) *Service {
	return &Service{
		repo:         repo,
		fileHandler:  fileHandler,
		categoryRepo: categoryRepo,
		clock:        clk,
		jwtSecret:    jwtSecret,
	}
}

// NewServiceWithEmail creates a new PPID management service with email support.
// This constructor allows optional email service injection for approval notifications.
func NewServiceWithEmail(repo Repository, fileHandler FileHandler, emailService EmailService, clk clock.Clock, jwtSecret string, emailConfig EmailConfig, categoryRepo CategoryLookup) *Service {
	return &Service{
		repo:         repo,
		fileHandler:  fileHandler,
		emailService: emailService,
		categoryRepo: categoryRepo,
		clock:        clk,
		jwtSecret:    jwtSecret,
		emailConfig:  emailConfig,
	}
}

// CreatePPIDInput represents the input for creating a PPID document
type CreatePPIDInput struct {
	Title         string
	Category      *string    // Optional
	Description   *string    // Optional
	PublicationAt *time.Time // Optional
	DocumentFile  io.Reader
	DocumentName  string
	DocumentSize  int64
	ContentType   string
	ThumbnailFile io.Reader // Optional - thumbnail image
	ThumbnailName string
	ThumbnailSize int64
	ThumbnailType string
}

// UpdatePPIDInput represents the input for updating a PPID document
type UpdatePPIDInput struct {
	Title              string
	Category           *string    // Optional
	Description        *string    // Optional
	PublicationAt      *time.Time // Optional - publication date
	ClearPublicationAt bool       // If true, clear publication_at (set to nil)
	DocumentFile       io.Reader  // Optional - only if updating document
	DocumentName       string
	DocumentSize       int64
	ContentType        string
	ThumbnailFile      io.Reader // Optional - thumbnail image
	ThumbnailName      string
	ThumbnailSize      int64
	ThumbnailType      string
}

// ListPPIDInput represents the input for listing PPID documents
type ListPPIDInput struct {
	Category *string // Optional filter by category
	Query    *string // Optional search query (title)
	Page     int
	Limit    int
}

// CreateRequestInput represents the input for creating a PPID request
type CreateRequestInput struct {
	PPIDId         string
	RequesterName  string
	RequesterEmail string
	Purpose        *string // Optional
}

// ListRequestsInput represents the input for listing PPID requests
type ListRequestsInput struct {
	Statuses []string // Optional filter by status (e.g., []string{"pending", "approved"})
	Page     int
	Limit    int
}

// Create creates a new PPID document with a document file.
// Returns concrete PPID struct.
//
// Validates: Requirements 12.5
func (s *Service) Create(ctx context.Context, input CreatePPIDInput) (*ppid.PPID, error) {
	// Validate the category exists, if one is provided. The FK is enforced
	// at the DB level, but checking here provides a friendlier 400 response.
	if input.Category != nil && *input.Category != "" {
		if _, err := s.categoryRepo.FindByID(ctx, *input.Category); err != nil {
			return nil, fmt.Errorf("invalid category: %w", err)
		}
	}

	// Save document file
	documentURL, err := s.fileHandler.SaveDocument(ctx, input.DocumentName, input.DocumentFile, input.DocumentSize, input.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to save document: %w", err)
	}

	// Save thumbnail file if provided
	var thumbnailURL *string
	if input.ThumbnailFile != nil {
		thumbURL, err := s.fileHandler.SaveImage(ctx, input.ThumbnailName, input.ThumbnailFile, input.ThumbnailSize, input.ThumbnailType)
		if err != nil {
			// Clean up document on thumbnail save failure
			_ = s.fileHandler.Delete(ctx, documentURL)
			return nil, fmt.Errorf("failed to save thumbnail: %w", err)
		}
		thumbnailURL = &thumbURL
	}

	// Create PPID entity
	now := s.clock.Now()
	p := &ppid.PPID{
		ID:            uuid.New().String(),
		Title:         input.Title,
		Category:      input.Category,
		DocumentURL:   &documentURL,
		ThumbnailURL:  thumbnailURL,
		Description:   input.Description,
		PublicationAt: input.PublicationAt,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Validate domain invariants
	if err := p.Validate(); err != nil {
		// Clean up uploaded files on validation failure
		_ = s.fileHandler.Delete(ctx, documentURL)
		if thumbnailURL != nil {
			_ = s.fileHandler.Delete(ctx, *thumbnailURL)
		}
		return nil, fmt.Errorf("invalid ppid data: %w", err)
	}

	// Persist PPID
	if err := s.repo.Create(ctx, p); err != nil {
		// Clean up uploaded files on persistence failure
		_ = s.fileHandler.Delete(ctx, documentURL)
		if thumbnailURL != nil {
			_ = s.fileHandler.Delete(ctx, *thumbnailURL)
		}
		return nil, fmt.Errorf("failed to create ppid: %w", err)
	}

	return p, nil
}

// GetByID retrieves a PPID document by ID.
// Returns concrete PPID struct.
//
// Validates: Requirements 12.4
func (s *Service) GetByID(ctx context.Context, id string) (*ppid.PPID, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ppid not found: %w", err)
	}

	return p, nil
}

// List retrieves paginated PPID documents with optional filtering.
// Returns concrete PPID structs and pagination metadata.
//
// Validates: Requirements 12.1, 12.2, 12.3
func (s *Service) List(ctx context.Context, input ListPPIDInput) ([]*ppid.PPID, pagination.Result, error) {
	// Validate and normalize pagination parameters
	offset, validatedLimit, err := pagination.Paginate(input.Page, input.Limit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("invalid pagination parameters: %w", err)
	}

	// Retrieve PPID documents from repository with filters
	ppidList, total, err := s.repo.List(ctx, input.Category, input.Query, offset, validatedLimit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("failed to list ppid: %w", err)
	}

	// Create pagination metadata
	paginationResult := pagination.NewResult(input.Page, validatedLimit, total)

	return ppidList, paginationResult, nil
}

// Update updates an existing PPID document.
// If DocumentFile is provided, updates the document; otherwise keeps existing document.
// Returns concrete PPID struct.
//
// Validates: Requirements 12.6
func (s *Service) Update(ctx context.Context, id string, input UpdatePPIDInput) (*ppid.PPID, error) {
	// Validate the new category exists, if one is provided.
	if input.Category != nil && *input.Category != "" {
		if _, err := s.categoryRepo.FindByID(ctx, *input.Category); err != nil {
			return nil, fmt.Errorf("invalid category: %w", err)
		}
	}

	// Retrieve existing PPID
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ppid not found: %w", err)
	}

	var oldDocumentURL *string
	if p.DocumentURL != nil {
		oldDocumentURL = new(string)
		*oldDocumentURL = *p.DocumentURL
	}

	var oldThumbnailURL *string
	if p.ThumbnailURL != nil {
		oldThumbnailURL = new(string)
		*oldThumbnailURL = *p.ThumbnailURL
	}

	// Update document if provided
	if input.DocumentFile != nil {
		newDocumentURL, err := s.fileHandler.SaveDocument(ctx, input.DocumentName, input.DocumentFile, input.DocumentSize, input.ContentType)
		if err != nil {
			return nil, fmt.Errorf("failed to save new document: %w", err)
		}
		p.DocumentURL = &newDocumentURL
	}

	// Update thumbnail if provided
	if input.ThumbnailFile != nil {
		newThumbnailURL, err := s.fileHandler.SaveImage(ctx, input.ThumbnailName, input.ThumbnailFile, input.ThumbnailSize, input.ThumbnailType)
		if err != nil {
			// Clean up new document if thumbnail save fails
			if input.DocumentFile != nil && p.DocumentURL != nil {
				_ = s.fileHandler.Delete(ctx, *p.DocumentURL)
			}
			return nil, fmt.Errorf("failed to save new thumbnail: %w", err)
		}
		p.ThumbnailURL = &newThumbnailURL
	}

	// Update fields
	p.Title = input.Title
	p.Category = input.Category
	p.Description = input.Description
	if input.ClearPublicationAt {
		p.PublicationAt = nil
	} else if input.PublicationAt != nil {
		p.PublicationAt = input.PublicationAt
	}
	p.UpdatedAt = s.clock.Now()

	// Validate domain invariants
	if err := p.Validate(); err != nil {
		// Clean up new files if validation fails
		if input.DocumentFile != nil && p.DocumentURL != nil {
			_ = s.fileHandler.Delete(ctx, *p.DocumentURL)
		}
		if input.ThumbnailFile != nil && p.ThumbnailURL != nil {
			_ = s.fileHandler.Delete(ctx, *p.ThumbnailURL)
		}
		return nil, fmt.Errorf("invalid ppid data: %w", err)
	}

	// Persist changes
	if err := s.repo.Update(ctx, p); err != nil {
		// Clean up new files if persistence fails
		if input.DocumentFile != nil && p.DocumentURL != nil {
			_ = s.fileHandler.Delete(ctx, *p.DocumentURL)
		}
		if input.ThumbnailFile != nil && p.ThumbnailURL != nil {
			_ = s.fileHandler.Delete(ctx, *p.ThumbnailURL)
		}
		return nil, fmt.Errorf("failed to update ppid: %w", err)
	}

	// Delete old document if a new one was uploaded successfully
	if input.DocumentFile != nil && oldDocumentURL != nil && *oldDocumentURL != "" {
		_ = s.fileHandler.Delete(ctx, *oldDocumentURL)
	}

	// Delete old thumbnail if a new one was uploaded successfully
	if input.ThumbnailFile != nil && oldThumbnailURL != nil && *oldThumbnailURL != "" {
		_ = s.fileHandler.Delete(ctx, *oldThumbnailURL)
	}

	return p, nil
}

// Delete removes a PPID document and its associated document file.
//
// Validates: Requirements 12.7
func (s *Service) Delete(ctx context.Context, id string) error {
	// Retrieve PPID to get document URL
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("ppid not found: %w", err)
	}

	// Delete PPID from database
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete ppid: %w", err)
	}

	// Delete associated document file if it exists
	if p.DocumentURL != nil && *p.DocumentURL != "" {
		if err := s.fileHandler.Delete(ctx, *p.DocumentURL); err != nil {
			// Log error but don't fail the operation since PPID is already deleted
			// In production, this should be logged properly
			_ = err
		}
	}

	return nil
}

// CreateRequest creates a new PPID document request.
// Returns concrete PPIDRequest struct.
//
// Validates: Requirements 12.8
func (s *Service) CreateRequest(ctx context.Context, input CreateRequestInput) (*ppid.PPIDRequest, error) {
	// Verify PPID document exists
	_, err := s.repo.FindByID(ctx, input.PPIDId)
	if err != nil {
		return nil, fmt.Errorf("ppid not found: %w", err)
	}

	// Create PPIDRequest entity
	now := s.clock.Now()
	r := &ppid.PPIDRequest{
		ID:             uuid.New().String(),
		PPIDId:         input.PPIDId,
		RequesterName:  input.RequesterName,
		RequesterEmail: input.RequesterEmail,
		Purpose:        input.Purpose,
		Status:         "pending",
		CreatedAt:      now,
	}

	// Validate domain invariants
	if err := r.Validate(); err != nil {
		return nil, fmt.Errorf("invalid ppid request data: %w", err)
	}

	// Persist request
	if err := s.repo.CreateRequest(ctx, r); err != nil {
		return nil, fmt.Errorf("failed to create ppid request: %w", err)
	}

	return r, nil
}

// ListRequests retrieves paginated PPID requests with optional status filtering.
// Returns concrete PPIDRequest structs and pagination metadata.
//
// Validates: Requirements 12.9
func (s *Service) ListRequests(ctx context.Context, input ListRequestsInput) ([]*ppid.PPIDRequest, pagination.Result, error) {
	// Validate and normalize pagination parameters
	offset, validatedLimit, err := pagination.Paginate(input.Page, input.Limit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("invalid pagination parameters: %w", err)
	}

	// Retrieve requests from repository with optional status filter
	requests, total, err := s.repo.ListRequests(ctx, input.Statuses, offset, validatedLimit)
	if err != nil {
		return nil, pagination.Result{}, fmt.Errorf("failed to list ppid requests: %w", err)
	}

	// Create pagination metadata
	paginationResult := pagination.NewResult(input.Page, validatedLimit, total)

	return requests, paginationResult, nil
}

// AccessTokenClaims represents JWT claims for document download
// Supports two types of tokens:
// 1. Document access token - from PPID request approval (document_id set)
// 2. User JWT - from login (user_id set) - allows admin to download any document
type AccessTokenClaims struct {
	DocumentId string `json:"document_id,omitempty"` // PPID document ID (for approved requests)
	UserID     string `json:"user_id,omitempty"`     // Admin user ID (for admin access)
	Email      string `json:"email,omitempty"`       // Admin email
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a JWT token for document access with 10-minute expiration
func (s *Service) GenerateAccessToken(ppidID string) (string, error) {
	now := s.clock.Now()
	expirationTime := now.Add(10 * time.Minute)

	claims := AccessTokenClaims{
		DocumentId: ppidID,
		UserID:     ppidID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateAccessToken validates and extracts claims from a JWT token
func (s *Service) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(s.clock.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

// UserClaims represents JWT claims from user login. The UserID is decoded
// via MapClaims to avoid the uuid.UUID vs string type mismatch in the
// underlying pkg/jwt.Claims struct.
type UserClaims struct {
	UserID string
	Email  string
	Exp    int64
}

// ValidateUserToken validates a user JWT token from login and returns claims
func (s *Service) ValidateUserToken(tokenString string) (*UserClaims, error) {
	mapClaims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, mapClaims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	c := &UserClaims{}
	if v, ok := mapClaims["user_id"].(string); ok {
		c.UserID = v
	}
	if v, ok := mapClaims["email"].(string); ok {
		c.Email = v
	}
	if v, ok := mapClaims["exp"].(float64); ok {
		c.Exp = int64(v)
	}
	if c.Exp > 0 && time.Unix(c.Exp, 0).Before(s.clock.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	return c, nil
}

// ApproveRequest approves a pending PPID request and generates an access token
func (s *Service) ApproveRequest(ctx context.Context, requestID string, adminUserID string) (*ppid.PPIDRequest, error) {
	// Retrieve the request
	req, err := s.repo.FindRequestByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("ppid request not found: %w", err)
	}

	// Validate request status is pending
	if !req.IsPending() {
		if req.IsApproved() {
			return nil, fmt.Errorf("request already approved")
		}
		if req.IsRevoked() {
			return nil, fmt.Errorf("cannot approve a revoked request")
		}
	}

	// Generate JWT access token
	token, err := s.GenerateAccessToken(req.PPIDId)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Get PPID document title (needed for email)
	ppidDoc, err := s.repo.FindByID(ctx, req.PPIDId)
	if err != nil {
		return nil, fmt.Errorf("failed to get PPID document: %w", err)
	}

	// Build download link with token
	downloadLink := fmt.Sprintf("%s/api/v1/ppid/document/%s/download?token=%s", s.emailConfig.DomainAddr, ppidDoc.ID, token)

	// Calculate token expiration time
	now := s.clock.Now()
	expirationTime := now.Add(10 * time.Minute).Format("2006-01-02T15:04:05Z07:00")

	// Send email FIRST before updating status
	// If email fails, status won't be updated
	emailInput := SendApprovalEmailInput{
		RequesterEmail:      req.RequesterEmail,
		RequesterName:       req.RequesterName,
		DocumentTitle:       ppidDoc.Title,
		DownloadLink:        downloadLink,
		TokenExpirationTime: expirationTime,
		VillageName:         s.emailConfig.VillageName,
		SupportEmail:        s.emailConfig.SupportEmail,
		WebsiteURL:          s.emailConfig.WebsiteURL,
	}

	if err := s.emailService.SendApprovalEmail(ctx, emailInput); err != nil {
		return nil, fmt.Errorf("failed to send approval email: %w", err)
	}

	// Update request status AFTER email is sent successfully
	req.Status = "approved"
	req.ApprovedAt = &now
	req.ApprovedBy = &adminUserID

	// Persist updated request
	if err := s.repo.UpdateRequest(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to update request: %w", err)
	}

	return req, nil
}

// RevokeRequest revokes a PPID request
func (s *Service) RevokeRequest(ctx context.Context, requestID string, adminUserID string) (*ppid.PPIDRequest, error) {
	// Retrieve the request
	req, err := s.repo.FindRequestByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("ppid request not found: %w", err)
	}

	// Validate request is not already revoked
	if req.IsRevoked() {
		return nil, fmt.Errorf("request already revoked")
	}

	// Update request status
	now := s.clock.Now()
	req.Status = "revoked"
	req.RevokedAt = &now
	req.RevokedBy = &adminUserID

	// Persist updated request
	if err := s.repo.UpdateRequest(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to update request: %w", err)
	}

	return req, nil
}

// DownloadDocumentInput represents input for document download authorization
type DownloadDocumentInput struct {
	DocumentID string
	Token      string
}

// DownloadDocumentResult represents the result of a successful document download authorization
type DownloadDocumentResult struct {
	FilePath    string
	ContentType string
	Filename    string
}

// DownloadDocument validates authorization and returns file info for download.
// It supports two token types:
// 1. Document access token (from PPID request approval) - validates document_id matches path
// 2. User JWT (from login) - allows admin to download any document
func (s *Service) DownloadDocument(ctx context.Context, input DownloadDocumentInput) (*DownloadDocumentResult, error) {
	if input.DocumentID == "" {
		return nil, fmt.Errorf("document ID is required")
	}
	if input.Token == "" {
		return nil, fmt.Errorf("access token required")
	}

	docClaims, docErr := s.ValidateAccessToken(input.Token)
	var userClaims *UserClaims
	var userErr error
	if docErr != nil || (docClaims != nil && docClaims.DocumentId == "") {
		userClaims, userErr = s.ValidateUserToken(input.Token)
	}

	var docIDToUse string
	var isAdmin bool

	if docErr == nil && docClaims != nil && docClaims.DocumentId != "" {
		if docClaims.DocumentId != input.DocumentID {
			return nil, fmt.Errorf("token not valid for this document")
		}
		docIDToUse = docClaims.DocumentId
	} else if userErr == nil && userClaims != nil {
		isAdmin = true
	} else {
		return nil, fmt.Errorf("invalid access token")
	}

	if docIDToUse == "" && !isAdmin {
		return nil, fmt.Errorf("invalid token: must have document_id or user_id")
	}

	if docIDToUse == "" {
		docIDToUse = input.DocumentID
	}

	ppid, err := s.repo.FindByID(ctx, docIDToUse)
	if err != nil {
		return nil, fmt.Errorf("document not found")
	}

	if ppid.DocumentURL == nil || *ppid.DocumentURL == "" {
		return nil, fmt.Errorf("document file not found")
	}

	filename := *ppid.DocumentURL
	filePath := s.fileHandler.GetFilePath(filename)

	contentType := s.getContentType(filename)

	return &DownloadDocumentResult{
		FilePath:    filePath,
		ContentType: contentType,
		Filename:    filename,
	}, nil
}

func (s *Service) getContentType(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(lower, ".gif"):
		return "image/gif"
	case strings.HasSuffix(lower, ".webp"):
		return "image/webp"
	default:
		return "application/pdf"
	}
}
