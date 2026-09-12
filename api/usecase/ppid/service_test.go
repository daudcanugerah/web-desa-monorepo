package ppid

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"webdesa/api/domain/ppid"
	"webdesa/api/domain/ppidcategory"
	galleryuc "webdesa/api/usecase/gallery"
)

// mockRepository is a mock implementation of Repository for testing
type mockRepository struct {
	ppidByID     map[string]*ppid.PPID
	requestByID  map[string]*ppid.PPIDRequest
	shouldFail   bool
	failureError error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		ppidByID:    make(map[string]*ppid.PPID),
		requestByID: make(map[string]*ppid.PPIDRequest),
	}
}

func (m *mockRepository) Create(ctx context.Context, p *ppid.PPID) error {
	if m.shouldFail {
		return m.failureError
	}
	m.ppidByID[p.ID] = p
	return nil
}

func (m *mockRepository) FindByID(ctx context.Context, id string) (*ppid.PPID, error) {
	if m.shouldFail {
		return nil, m.failureError
	}
	if p, ok := m.ppidByID[id]; ok {
		return p, nil
	}
	return nil, errors.New("ppid not found")
}

func (m *mockRepository) List(ctx context.Context, category *string, query *string, offset, limit int) ([]*ppid.PPID, int, error) {
	if m.shouldFail {
		return nil, 0, m.failureError
	}
	return []*ppid.PPID{}, 0, nil
}

func (m *mockRepository) Update(ctx context.Context, p *ppid.PPID) error {
	if m.shouldFail {
		return m.failureError
	}
	m.ppidByID[p.ID] = p
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	if m.shouldFail {
		return m.failureError
	}
	delete(m.ppidByID, id)
	return nil
}

func (m *mockRepository) CreateRequest(ctx context.Context, r *ppid.PPIDRequest) error {
	if m.shouldFail {
		return m.failureError
	}
	m.requestByID[r.ID] = r
	return nil
}

func (m *mockRepository) ListRequests(ctx context.Context, statuses []string, offset, limit int) ([]*ppid.PPIDRequest, int, error) {
	if m.shouldFail {
		return nil, 0, m.failureError
	}
	return []*ppid.PPIDRequest{}, 0, nil
}

func (m *mockRepository) GetCategories(ctx context.Context) ([]string, error) {
	if m.shouldFail {
		return nil, m.failureError
	}
	return []string{}, nil
}

func (m *mockRepository) UpdateRequest(ctx context.Context, r *ppid.PPIDRequest) error {
	if m.shouldFail {
		return m.failureError
	}
	m.requestByID[r.ID] = r
	return nil
}

func (m *mockRepository) FindRequestByID(ctx context.Context, id string) (*ppid.PPIDRequest, error) {
	if m.shouldFail {
		return nil, m.failureError
	}
	if r, ok := m.requestByID[id]; ok {
		return r, nil
	}
	return nil, errors.New("ppid request not found")
}

// mockFileStore is a mock implementation of gallery.FileStore for testing.
type mockFileStore struct{}

func (m *mockFileStore) SaveImage(ctx context.Context, feature string, in galleryuc.FileInput) (galleryuc.SavedFile, error) {
	return galleryuc.SavedFile{MediaID: "mock-media-id"}, nil
}

func (m *mockFileStore) SaveDocument(ctx context.Context, feature string, in galleryuc.FileInput) (galleryuc.SavedFile, error) {
	return galleryuc.SavedFile{MediaID: "mock-media-id"}, nil
}

func (m *mockFileStore) SaveImages(ctx context.Context, feature string, inputs []galleryuc.FileInput) ([]galleryuc.SavedFile, []error) {
	out := make([]galleryuc.SavedFile, len(inputs))
	for i := range inputs {
		out[i] = galleryuc.SavedFile{MediaID: "mock-media-id"}
	}
	return out, nil
}

func (m *mockFileStore) ValidateFiles(ctx context.Context, feature string, inputs []galleryuc.FileInput) []error {
	return make([]error, len(inputs))
}

func (m *mockFileStore) Delete(ctx context.Context, mediaID string) error {
	return nil
}

func (m *mockFileStore) Open(ctx context.Context, mediaID string) (*galleryuc.MediaBinary, error) {
	return &galleryuc.MediaBinary{
		FilePath:    "/mock/path/doc.pdf",
		ContentType: "application/pdf",
		Filename:    mediaID,
		Size:        128,
	}, nil
}

// mockClock is a mock implementation of Clock for testing
type mockClock struct {
	now time.Time
}

func (m *mockClock) Now() time.Time {
	return m.now
}

// mockCategoryLookup is a no-op mock for CategoryLookup.
// It returns success for any id, which is sufficient for tests that don't
// exercise the category validation path.
type mockCategoryLookup struct{}

func (m *mockCategoryLookup) FindByID(ctx context.Context, id string) (*ppidcategory.Category, error) {
	return &ppidcategory.Category{ID: id, Name: "mock"}, nil
}

// mockEmailService is a mock implementation of EmailService for testing
type mockEmailService struct {
	shouldFail bool
	failError  error
}

func (m *mockEmailService) SendApprovalEmail(ctx context.Context, input SendApprovalEmailInput) error {
	if m.shouldFail {
		return m.failError
	}
	return nil
}

// TestGenerateAccessToken tests JWT token generation
func TestGenerateAccessToken(t *testing.T) {
	repo := newMockRepository()
	fileStore := &mockFileStore{}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	service := NewService(repo, fileStore, clk, jwtSecret, &mockCategoryLookup{})

	ppidID := "ppid-123"

	token, err := service.GenerateAccessToken(ppidID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the token
	claims, err := service.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, ppidID, claims.DocumentId)
	assert.Equal(t, ppidID, claims.UserID)
	assert.NotEmpty(t, claims.IssuedAt)
	assert.NotEmpty(t, claims.ExpiresAt)
}

// TestGenerateAccessTokenExpiration tests that tokens expire after 10 minutes
func TestGenerateAccessTokenExpiration(t *testing.T) {
	repo := newMockRepository()
	fileStore := &mockFileStore{}
	now := time.Now()
	clk := &mockClock{now: now}
	jwtSecret := "test-secret-key"
	service := NewService(repo, fileStore, clk, jwtSecret, &mockCategoryLookup{})

	token, err := service.GenerateAccessToken("ppid-123")
	require.NoError(t, err)

	// Move time forward to expire the token
	clk.now = now.Add(11 * time.Minute)

	// Token should be expired
	_, err = service.ValidateAccessToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

// TestValidateAccessToken tests token validation
func TestValidateAccessToken(t *testing.T) {
	repo := newMockRepository()
	fileStore := &mockFileStore{}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	service := NewService(repo, fileStore, clk, jwtSecret, &mockCategoryLookup{})

	// Generate a valid token
	token, err := service.GenerateAccessToken("ppid-123")
	require.NoError(t, err)

	// Valid token should pass
	claims, err := service.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.NotNil(t, claims)

	// Invalid token should fail
	_, err = service.ValidateAccessToken("invalid-token")
	assert.Error(t, err)

	// Token with wrong secret should fail
	service2 := NewService(repo, fileStore, clk, "different-secret", &mockCategoryLookup{})
	_, err = service2.ValidateAccessToken(token)
	assert.Error(t, err)
}

// TestApproveRequest tests approving a pending request
func TestApproveRequest(t *testing.T) {
	repo := newMockRepository()
	fileStore := &mockFileStore{}
	emailService := &mockEmailService{}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	emailConfig := EmailConfig{
		VillageName:  "Test Village",
		SupportEmail: "support@test.com",
		WebsiteURL:   "http://test.com",
	}
	service := NewServiceWithEmail(repo, fileStore, emailService, clk, jwtSecret, emailConfig, &mockCategoryLookup{})

	// Create test PPID and request
	ppidID := "ppid-123"
	requestID := "req-123"
	adminUserID := "admin-123"

	ppidDoc := &ppid.PPID{
		ID:        ppidID,
		Title:     "Test Document",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.ppidByID[ppidID] = ppidDoc

	request := &ppid.PPIDRequest{
		ID:             requestID,
		PPIDId:         ppidID,
		RequesterName:  "John Doe",
		RequesterEmail: "john@example.com",
		Status:         "pending",
		CreatedAt:      time.Now(),
	}
	repo.requestByID[requestID] = request

	// Approve the request
	approvedReq, err := service.ApproveRequest(context.Background(), requestID, adminUserID)
	require.NoError(t, err)
	assert.NotNil(t, approvedReq)
	assert.Equal(t, "approved", approvedReq.Status)
	assert.NotNil(t, approvedReq.ApprovedAt)
	assert.Equal(t, adminUserID, *approvedReq.ApprovedBy)
}

// TestApproveRequestEmailFailKeepsPending locks the documented approval
// semantics (feature-docs/ppid/concept.md): the approval email must be
// delivered before the row flips to approved. When SMTP fails the
// operator gets an error and the request stays pending so approve can
// simply be retried — it must never end up approved without an email.
func TestApproveRequestEmailFailKeepsPending(t *testing.T) {
	repo := newMockRepository()
	fileStore := &mockFileStore{}
	emailService := &mockEmailService{shouldFail: true, failError: errors.New("smtp connection refused")}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	emailConfig := EmailConfig{
		VillageName:  "Test Village",
		SupportEmail: "support@test.com",
		WebsiteURL:   "http://test.com",
	}
	service := NewServiceWithEmail(repo, fileStore, emailService, clk, jwtSecret, emailConfig, &mockCategoryLookup{})

	ppidID := "ppid-mailfail"
	requestID := "req-mailfail"
	adminUserID := "admin-123"

	repo.ppidByID[ppidID] = &ppid.PPID{ID: ppidID, Title: "Test Document", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	repo.requestByID[requestID] = &ppid.PPIDRequest{
		ID:             requestID,
		PPIDId:         ppidID,
		RequesterName:  "John Doe",
		RequesterEmail: "john@example.com",
		Status:         "pending",
		CreatedAt:      time.Now(),
	}

	approvedReq, err := service.ApproveRequest(context.Background(), requestID, adminUserID)
	require.Error(t, err)
	assert.Nil(t, approvedReq)
	assert.Contains(t, err.Error(), "failed to send approval email")

	// The row must still be pending in the repository.
	stored, ok := repo.requestByID[requestID]
	require.True(t, ok)
	assert.Equal(t, "pending", stored.Status)
	assert.Nil(t, stored.ApprovedAt)
	assert.Nil(t, stored.ApprovedBy)
}

// TestApproveRequestWithSignedURL exercises the Task 7.2 path: the
// approve handler mints a JWT scoped to (scope=ppid, sub=ppid_request:<id>,
// media_id=document_media_id) and returns a download link targeting the
// unified /api/v1/media/{id}/...?jwt= route. Revoking the request adds
// the sub to the deny list so subsequent token uses fail with
// ErrInvalidSignedToken.
func TestApproveRequestWithSignedURL(t *testing.T) {
	repo := newMockRepository()
	fileStore := &mockFileStore{}
	emailService := &mockEmailService{}
	clk := &mockClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	jwtSecret := "test-secret-key"
	emailConfig := EmailConfig{
		VillageName:  "Test Village",
		SupportEmail: "support@test.com",
		WebsiteURL:   "http://test.com",
		DomainAddr:   "https://test.com",
	}
	signer := galleryuc.NewSignedURLService(jwtSecret, clk.Now)
	service := NewServiceWithSignedURL(repo, fileStore, emailService, clk, jwtSecret, emailConfig, &mockCategoryLookup{}, signer)

	requestID := "req-456"
	adminUserID := "admin-456"
	mediaID := "media-abc"

	docID := "ppid-doc-uuid"
	repo.ppidByID[docID] = &ppid.PPID{
		ID:              docID,
		Title:           "Test Document",
		DocumentMediaID: &mediaID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	request := &ppid.PPIDRequest{
		ID:             requestID,
		PPIDId:         docID,
		RequesterName:  "Jane",
		RequesterEmail: "jane@example.com",
		Status:         "pending",
		CreatedAt:      time.Now(),
	}
	repo.requestByID[requestID] = request

	approved, err := service.ApproveRequest(context.Background(), requestID, adminUserID)
	require.NoError(t, err)
	assert.Equal(t, "approved", approved.Status)

	// The email should have received the signed download URL (the exact
	// URL contents are not asserted here — see the handler tests for that).
}

// TestRevokeDeniesSignedToken checks Task 7.2's revocation flow: after
// RevokeRequest runs, a previously valid token must fail Verify.
func TestRevokeDeniesSignedToken(t *testing.T) {
	clk := &mockClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	signer := galleryuc.NewSignedURLService("test-secret-key", clk.Now)
	repo := newMockRepository()
	fileStore := &mockFileStore{}
	emailService := &mockEmailService{}
	emailConfig := EmailConfig{DomainAddr: "https://test.com"}
	service := NewServiceWithSignedURL(repo, fileStore, emailService, clk, "test-secret-key", emailConfig, &mockCategoryLookup{}, signer)

	docID := "ppid-doc-uuid-2"
	mediaID := "media-def"
	repo.ppidByID[docID] = &ppid.PPID{
		ID:              docID,
		Title:           "Test",
		DocumentMediaID: &mediaID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	requestID := "req-revoke"
	repo.requestByID[requestID] = &ppid.PPIDRequest{
		ID:     requestID,
		PPIDId: docID,
		Status: "pending",
	}

	_, err := service.ApproveRequest(context.Background(), requestID, "admin-x")
	require.NoError(t, err)

	// Pre-revoke: the token signs/verifies cleanly.
	tok, err := signer.Sign(galleryuc.ScopePPID, mediaID, "ppid_request:"+requestID, 0)
	require.NoError(t, err)
	claims, err := signer.Verify(tok, mediaID)
	require.NoError(t, err)
	assert.Equal(t, galleryuc.ScopePPID, claims.Scope)

	// Revoke — token must now fail Verify.
	_, err = service.RevokeRequest(context.Background(), requestID, "admin-x")
	require.NoError(t, err)
	_, err = signer.Verify(tok, mediaID)
	assert.ErrorIs(t, err, galleryuc.ErrInvalidSignedToken)
}

// TestApproveRequestNotFound tests approving a non-existent request
func TestApproveRequestNotFound(t *testing.T) {
	repo := newMockRepository()
	fileStore := &mockFileStore{}
	emailService := &mockEmailService{}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	emailConfig := EmailConfig{
		VillageName:  "Test Village",
		SupportEmail: "support@test.com",
		WebsiteURL:   "http://test.com",
	}
	service := NewServiceWithEmail(repo, fileStore, emailService, clk, jwtSecret, emailConfig, &mockCategoryLookup{})

	_, err := service.ApproveRequest(context.Background(), "non-existent", "admin-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestApproveRequestAlreadyApproved tests that re-approving fails
func TestApproveRequestAlreadyApproved(t *testing.T) {
	repo := newMockRepository()
	fileStore := &mockFileStore{}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	service := NewService(repo, fileStore, clk, jwtSecret, &mockCategoryLookup{})

	requestID := "req-123"
	adminUserID := "admin-123"
	now := time.Now()

	request := &ppid.PPIDRequest{
		ID:             requestID,
		PPIDId:         "ppid-123",
		RequesterName:  "John Doe",
		RequesterEmail: "john@example.com",
		Status:         "approved",
		ApprovedAt:     &now,
		ApprovedBy:     &adminUserID,
		CreatedAt:      time.Now(),
	}
	repo.requestByID[requestID] = request

	_, err := service.ApproveRequest(context.Background(), requestID, "another-admin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already approved")
}

// TestRevokeRequest tests revoking a request
func TestRevokeRequest(t *testing.T) {
	repo := newMockRepository()
	fileStore := &mockFileStore{}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	service := NewService(repo, fileStore, clk, jwtSecret, &mockCategoryLookup{})

	requestID := "req-123"
	adminUserID := "admin-123"

	request := &ppid.PPIDRequest{
		ID:             requestID,
		PPIDId:         "ppid-123",
		RequesterName:  "John Doe",
		RequesterEmail: "john@example.com",
		Status:         "pending",
		CreatedAt:      time.Now(),
	}
	repo.requestByID[requestID] = request

	// Revoke the request
	revokedReq, err := service.RevokeRequest(context.Background(), requestID, adminUserID)
	require.NoError(t, err)
	assert.NotNil(t, revokedReq)
	assert.Equal(t, "revoked", revokedReq.Status)
	assert.NotNil(t, revokedReq.RevokedAt)
	assert.Equal(t, adminUserID, *revokedReq.RevokedBy)
}

// TestRevokeRequestNotFound tests revoking a non-existent request
func TestRevokeRequestNotFound(t *testing.T) {
	repo := newMockRepository()
	fileStore := &mockFileStore{}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	service := NewService(repo, fileStore, clk, jwtSecret, &mockCategoryLookup{})

	_, err := service.RevokeRequest(context.Background(), "non-existent", "admin-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestRevokeRequestAlreadyRevoked tests that re-revoking fails
func TestRevokeRequestAlreadyRevoked(t *testing.T) {
	repo := newMockRepository()
	fileStore := &mockFileStore{}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	service := NewService(repo, fileStore, clk, jwtSecret, &mockCategoryLookup{})

	requestID := "req-123"
	adminUserID := "admin-123"
	now := time.Now()

	request := &ppid.PPIDRequest{
		ID:             requestID,
		PPIDId:         "ppid-123",
		RequesterName:  "John Doe",
		RequesterEmail: "john@example.com",
		Status:         "revoked",
		RevokedAt:      &now,
		RevokedBy:      &adminUserID,
		CreatedAt:      time.Now(),
	}
	repo.requestByID[requestID] = request

	_, err := service.RevokeRequest(context.Background(), requestID, "another-admin")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already revoked")
}

// TestTokenUniqueness tests that different tokens have different jti claims
func TestTokenUniqueness(t *testing.T) {
	repo := newMockRepository()
	fileStore := &mockFileStore{}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	service := NewService(repo, fileStore, clk, jwtSecret, &mockCategoryLookup{})

	token1, err := service.GenerateAccessToken("ppid-123")
	require.NoError(t, err)

	token2, err := service.GenerateAccessToken("ppid-123")
	require.NoError(t, err)

	// Tokens should be different
	assert.NotEqual(t, token1, token2)

	// Extract jti claims
	claims1, err := service.ValidateAccessToken(token1)
	require.NoError(t, err)

	claims2, err := service.ValidateAccessToken(token2)
	require.NoError(t, err)

	// JTI should be different
	assert.NotEqual(t, claims1.ID, claims2.ID)
}
