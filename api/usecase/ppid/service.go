package ppid

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"webdesa/api/domain/ppid"
	galleryUsecase "webdesa/api/usecase/gallery"
	"webdesa/api/pkg/clock"
	"webdesa/api/pkg/pagination"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Service implements PPID (public information disclosure) management business logic.
// It accepts interfaces (Repository, FileStore, EmailService) and returns concrete structs (PPID, PPIDRequest).
// This follows the "accept interfaces, return structs" Go idiom.
type Service struct {
	repo         Repository
	fileStore    galleryUsecase.FileStore
	emailService EmailService
	categoryRepo CategoryLookup
	clock        clock.Clock
	jwtSecret    string
	emailConfig  EmailConfig
	// signedURL (Task 7.2) backs the unified PPID download URL with a
	// JWT scoped to the request id, so revocation propagates instantly
	// via the gallery's DenyList. Optional: when nil, the service falls
	// back to the legacy GenerateAccessToken path and the old
	// /api/v1/ppid/document/{id}/download?token= route continues to work.
	signedURL *galleryUsecase.SignedURLService
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
func NewService(repo Repository, fileStore galleryUsecase.FileStore, clk clock.Clock, jwtSecret string, categoryRepo CategoryLookup) *Service {
	return &Service{
		repo:         repo,
		fileStore:    fileStore,
		categoryRepo: categoryRepo,
		clock:        clk,
		jwtSecret:    jwtSecret,
	}
}

// NewServiceWithSignedURL creates a PPID service backed by the unified
// SignedURLService (Task 7.2). Prefer this constructor in cmd/serve.go
// and integration tests; nil signedURL is allowed for the deprecation
// window.
func NewServiceWithSignedURL(repo Repository, fileStore galleryUsecase.FileStore, emailService EmailService, clk clock.Clock, jwtSecret string, emailConfig EmailConfig, categoryRepo CategoryLookup, signedURL *galleryUsecase.SignedURLService) *Service {
	return &Service{
		repo:         repo,
		fileStore:    fileStore,
		emailService: emailService,
		categoryRepo: categoryRepo,
		clock:        clk,
		jwtSecret:    jwtSecret,
		emailConfig:  emailConfig,
		signedURL:    signedURL,
	}
}

// NewServiceWithEmail creates a new PPID management service with email support.
// This constructor allows optional email service injection for approval notifications.
func NewServiceWithEmail(repo Repository, fileStore galleryUsecase.FileStore, emailService EmailService, clk clock.Clock, jwtSecret string, emailConfig EmailConfig, categoryRepo CategoryLookup) *Service {
	return &Service{
		repo:         repo,
		fileStore:    fileStore,
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
	// DocumentMediaID references a document already uploaded via
	// POST /ppid/upload-media. Mutually exclusive with DocumentFile.
	DocumentMediaID *string
	// ThumbnailMediaID references a thumbnail already uploaded via
	// POST /ppid/upload-thumbnail. Mutually exclusive with ThumbnailFile.
	ThumbnailMediaID *string
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
	// DocumentMediaID references a document already uploaded via
	// POST /ppid/upload-media. Mutually exclusive with DocumentFile.
	DocumentMediaID *string
	// ThumbnailMediaID references a thumbnail already uploaded via
	// POST /ppid/upload-thumbnail. Mutually exclusive with ThumbnailFile.
	ThumbnailMediaID *string
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

// resolveMediaRef returns the media id for a pre-uploaded media ref or a
// newly saved file. The second return tells the caller whether the media
// was created here (only that case may delete it on failure).
func (s *Service) resolveMediaRef(ctx context.Context, mediaRef *string, file io.Reader, saveFile func() (galleryUsecase.SavedFile, error)) (string, bool, error) {
	if mediaRef != nil && *mediaRef != "" {
		if file != nil {
			return "", false, errors.New("provide either a media id or a file, not both")
		}
		if _, err := s.fileStore.Open(ctx, *mediaRef); err != nil {
			return "", false, fmt.Errorf("invalid media id: %w", err)
		}
		return *mediaRef, false, nil
	}
	if file != nil {
		saved, err := saveFile()
		if err != nil {
			return "", false, err
		}
		return saved.MediaID, true, nil
	}
	return "", false, nil
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

	// Save document file via the gallery FileStore. The returned MediaID is
	// the canonical handle for the uploaded document.
	docMediaID, docCreated, err := s.resolveMediaRef(ctx, input.DocumentMediaID, input.DocumentFile, func() (galleryUsecase.SavedFile, error) {
		return s.fileStore.SaveDocument(ctx, galleryUsecase.FeaturePPID, galleryUsecase.FileInput{
			OriginalName: input.DocumentName,
			Content:      input.DocumentFile,
			Size:         input.DocumentSize,
			ContentType:  input.ContentType,
		})
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve document: %w", err)
	}
	if docMediaID == "" {
		return nil, fmt.Errorf("document file or media id is required")
	}

	// Save thumbnail file if provided
	thumbMediaID, thumbCreated, err := s.resolveMediaRef(ctx, input.ThumbnailMediaID, input.ThumbnailFile, func() (galleryUsecase.SavedFile, error) {
		return s.fileStore.SaveImage(ctx, galleryUsecase.FeaturePPID, galleryUsecase.FileInput{
			OriginalName: input.ThumbnailName,
			Content:      input.ThumbnailFile,
			Size:         input.ThumbnailSize,
			ContentType:  input.ThumbnailType,
		})
	})
	if err != nil {
		if docCreated {
			_ = s.fileStore.Delete(ctx, docMediaID)
		}
		return nil, fmt.Errorf("failed to resolve thumbnail: %w", err)
	}
	var thumbnailMediaID *string
	if thumbMediaID != "" {
		thumbnailMediaID = &thumbMediaID
	}

	// Create PPID entity
	now := s.clock.Now()
	p := &ppid.PPID{
		ID:               uuid.New().String(),
		Title:            input.Title,
		Category:         input.Category,
		DocumentMediaID:  &docMediaID,
		ThumbnailMediaID: thumbnailMediaID,
		Description:      input.Description,
		PublicationAt:    input.PublicationAt,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Validate domain invariants
	if err := p.Validate(); err != nil {
		// Clean up files created here on validation failure
		if docCreated {
			_ = s.fileStore.Delete(ctx, docMediaID)
		}
		if thumbCreated {
			_ = s.fileStore.Delete(ctx, thumbMediaID)
		}
		return nil, fmt.Errorf("invalid ppid data: %w", err)
	}

	// Persist PPID
	if err := s.repo.Create(ctx, p); err != nil {
		// Clean up files created here on persistence failure
		if docCreated {
			_ = s.fileStore.Delete(ctx, docMediaID)
		}
		if thumbCreated {
			_ = s.fileStore.Delete(ctx, thumbMediaID)
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

	var oldDocumentMediaID *string
	if p.DocumentMediaID != nil {
		oldDocumentMediaID = new(string)
		*oldDocumentMediaID = *p.DocumentMediaID
	}

	var oldThumbnailMediaID *string
	if p.ThumbnailMediaID != nil {
		oldThumbnailMediaID = new(string)
		*oldThumbnailMediaID = *p.ThumbnailMediaID
	}

	// Update document if provided
	docCreated := false
	if input.DocumentMediaID != nil || input.DocumentFile != nil {
		newDocID, created, err := s.resolveMediaRef(ctx, input.DocumentMediaID, input.DocumentFile, func() (galleryUsecase.SavedFile, error) {
			return s.fileStore.SaveDocument(ctx, galleryUsecase.FeaturePPID, galleryUsecase.FileInput{
				OriginalName: input.DocumentName,
				Content:      input.DocumentFile,
				Size:         input.DocumentSize,
				ContentType:  input.ContentType,
			})
		})
		if err != nil {
			return nil, fmt.Errorf("failed to resolve new document: %w", err)
		}
		docCreated = created
		p.DocumentMediaID = &newDocID
	}

	// Update thumbnail if provided
	thumbCreated := false
	if input.ThumbnailMediaID != nil || input.ThumbnailFile != nil {
		newThumbID, created, err := s.resolveMediaRef(ctx, input.ThumbnailMediaID, input.ThumbnailFile, func() (galleryUsecase.SavedFile, error) {
			return s.fileStore.SaveImage(ctx, galleryUsecase.FeaturePPID, galleryUsecase.FileInput{
				OriginalName: input.ThumbnailName,
				Content:      input.ThumbnailFile,
				Size:         input.ThumbnailSize,
				ContentType:  input.ThumbnailType,
			})
		})
		if err != nil {
			// Clean up new document if thumbnail resolution fails
			if docCreated && p.DocumentMediaID != nil {
				_ = s.fileStore.Delete(ctx, *p.DocumentMediaID)
			}
			return nil, fmt.Errorf("failed to resolve new thumbnail: %w", err)
		}
		thumbCreated = created
		p.ThumbnailMediaID = &newThumbID
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
		// Clean up files created here on validation failure
		if docCreated && p.DocumentMediaID != nil {
			_ = s.fileStore.Delete(ctx, *p.DocumentMediaID)
		}
		if thumbCreated && p.ThumbnailMediaID != nil {
			_ = s.fileStore.Delete(ctx, *p.ThumbnailMediaID)
		}
		return nil, fmt.Errorf("invalid ppid data: %w", err)
	}

	// Persist changes
	if err := s.repo.Update(ctx, p); err != nil {
		// Clean up files created here on persistence failure
		if docCreated && p.DocumentMediaID != nil {
			_ = s.fileStore.Delete(ctx, *p.DocumentMediaID)
		}
		if thumbCreated && p.ThumbnailMediaID != nil {
			_ = s.fileStore.Delete(ctx, *p.ThumbnailMediaID)
		}
		return nil, fmt.Errorf("failed to update ppid: %w", err)
	}

	// Delete old document if a new one was uploaded successfully
	if (input.DocumentFile != nil || (input.DocumentMediaID != nil && *input.DocumentMediaID != "")) &&
		oldDocumentMediaID != nil && *oldDocumentMediaID != "" {
		_ = s.fileStore.Delete(ctx, *oldDocumentMediaID)
	}

	// Delete old thumbnail if a new one was uploaded successfully
	if (input.ThumbnailFile != nil || (input.ThumbnailMediaID != nil && *input.ThumbnailMediaID != "")) &&
		oldThumbnailMediaID != nil && *oldThumbnailMediaID != "" {
		_ = s.fileStore.Delete(ctx, *oldThumbnailMediaID)
	}

	return p, nil
}

// Delete removes a PPID document and its associated document and thumbnail
// media in the gallery.
//
// Validates: Requirements 12.7
func (s *Service) Delete(ctx context.Context, id string) error {
	// Retrieve PPID to get media IDs
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("ppid not found: %w", err)
	}

	// Delete PPID from database
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete ppid: %w", err)
	}

	// Delete associated document media if it exists
	if p.DocumentMediaID != nil && *p.DocumentMediaID != "" {
		_ = s.fileStore.Delete(ctx, *p.DocumentMediaID)
	}

	// Delete associated thumbnail media if it exists
	if p.ThumbnailMediaID != nil && *p.ThumbnailMediaID != "" {
		_ = s.fileStore.Delete(ctx, *p.ThumbnailMediaID)
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

	// Get PPID document (needed for token + email)
	ppidDoc, err := s.repo.FindByID(ctx, req.PPIDId)
	if err != nil {
		return nil, fmt.Errorf("failed to get PPID document: %w", err)
	}

	// Build download token + URL.
	//
	// Task 7.2 path: when SignedURLService is wired, mint a JWT scoped
	// to (scope=ppid, sub=ppid_request:<id>, media_id=document's media id).
	// The unified /api/v1/media/{id}/...?jwt= handler serves the file,
	// and RevokeRequest can deny-list the sub to invalidate instantly.
	//
	// Legacy path: when SignedURLService is nil (deprecation window or
	// older tests), fall back to GenerateAccessToken + the old route.
	var (
		downloadLink    string
		expirationTime  string
		tokenExpiresAt  time.Time
	)
	if s.signedURL != nil {
		if ppidDoc.DocumentMediaID == nil || *ppidDoc.DocumentMediaID == "" {
			return nil, fmt.Errorf("ppid document has no media id; cannot mint signed URL")
		}
		t, err := s.signedURL.Sign(galleryUsecase.ScopePPID, *ppidDoc.DocumentMediaID, "ppid_request:"+req.ID, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to sign media token: %w", err)
		}
		downloadLink = fmt.Sprintf("%s%s?jwt=%s", s.emailConfig.DomainAddr, galleryUsecase.SignedURLPath("content", *ppidDoc.DocumentMediaID), url.QueryEscape(t))
	} else {
		t, err := s.GenerateAccessToken(req.PPIDId)
		if err != nil {
			return nil, fmt.Errorf("failed to generate access token: %w", err)
		}
		downloadLink = fmt.Sprintf("%s/api/v1/ppid/document/%s/download?token=%s", s.emailConfig.DomainAddr, ppidDoc.ID, t)
	}
	now := s.clock.Now()
	tokenExpiresAt = now.Add(1 * time.Hour)
	expirationTime = tokenExpiresAt.Format("2006-01-02T15:04:05Z07:00")

	// Persist the approval BEFORE sending the email so the database is
	// never out of sync with reality: if the email fails after this
	// point the request is still marked approved and the operator can
	// resend. If we sent the email first (previous behavior) and the DB
	// update then failed, the requester would receive a download link for
	// a request that the system still shows as pending.
	req.Status = "approved"
	req.ApprovedAt = &now
	req.ApprovedBy = &adminUserID

	if err := s.repo.UpdateRequest(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to update request: %w", err)
	}

	// Try to send the approval email. A failure here leaves the DB in
	// the approved state (above) and surfaces to the operator; the
	// requester simply does not receive the download link until the
	// operator resends (TODO: add /ppid/requests/{id}/resend-email).
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
		return nil, fmt.Errorf("request approved but email send failed (resend required): %w", err)
	}

	return req, nil
}

// RevokeRequest revokes a PPID request. When the service is wired with
// SignedURLService (Task 7.2), the token's sub is added to the gallery
// deny list so any future attempt to use the approved download URL
// returns 401 instantly.
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

	// Invalidate any outstanding signed-URL tokens for this request.
	if s.signedURL != nil {
		s.signedURL.DenyList().Revoke("ppid_request:" + req.ID)
	}

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

// DownloadDocumentResult represents the result of a successful document
// download authorization. ContentURL is the absolute path (including the
// API version prefix) the HTTP handler should redirect the client to; the
// gallery admin endpoint streams the underlying bytes.
type DownloadDocumentResult struct {
	MediaID     string
	ContentType string
	Filename    string
	Size        int64
	Binary      *galleryUsecase.MediaBinary
}

// DownloadDocument validates authorization and returns the document binary so
// the HTTP handler can stream the bytes directly. Grants (or admin tokens)
// are checked here; the handler never contacts the gallery serving layer.
//
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

	ppidDoc, err := s.repo.FindByID(ctx, docIDToUse)
	if err != nil {
		return nil, fmt.Errorf("document not found")
	}

	if ppidDoc.DocumentMediaID == nil || *ppidDoc.DocumentMediaID == "" {
		return nil, fmt.Errorf("document file not found")
	}

	mediaID := *ppidDoc.DocumentMediaID
	bin, err := s.fileStore.Open(ctx, mediaID)
	if err != nil {
		return nil, fmt.Errorf("document file not found")
	}

	// Filename hint for the Content-Disposition header.
	filename := mediaID
	contentType := bin.ContentType
	if contentType == "" {
		contentType = "application/pdf"
	}

	return &DownloadDocumentResult{
		MediaID:     mediaID,
		ContentType: contentType,
		Filename:    filename,
		Size:        bin.Size,
		Binary:      bin,
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
