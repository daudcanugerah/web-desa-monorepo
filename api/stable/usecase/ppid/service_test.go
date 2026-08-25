package ppid

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"webdesa/api/domain/ppid"
	"webdesa/api/domain/ppidcategory"
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

// mockFileHandler is a mock implementation of FileHandler for testing
type mockFileHandler struct{}

func (m *mockFileHandler) SaveDocument(ctx context.Context, name string, file io.Reader, size int64, contentType string) (string, error) {
	return "mock-url.pdf", nil
}

func (m *mockFileHandler) SaveImage(ctx context.Context, name string, file io.Reader, size int64, contentType string) (string, error) {
	return "mock-url.png", nil
}

func (m *mockFileHandler) Delete(ctx context.Context, url string) error {
	return nil
}

func (m *mockFileHandler) GetFile(ctx context.Context, filename string) ([]byte, string, error) {
	return []byte("mock-content"), "application/pdf", nil
}

func (m *mockFileHandler) GetFilePath(relativePath string) string {
	return "/mock/uploads/" + relativePath
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
	fileHandler := &mockFileHandler{}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	service := NewService(repo, fileHandler, clk, jwtSecret, &mockCategoryLookup{})

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
	fileHandler := &mockFileHandler{}
	now := time.Now()
	clk := &mockClock{now: now}
	jwtSecret := "test-secret-key"
	service := NewService(repo, fileHandler, clk, jwtSecret, &mockCategoryLookup{})

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
	fileHandler := &mockFileHandler{}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	service := NewService(repo, fileHandler, clk, jwtSecret, &mockCategoryLookup{})

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
	service2 := NewService(repo, fileHandler, clk, "different-secret", &mockCategoryLookup{})
	_, err = service2.ValidateAccessToken(token)
	assert.Error(t, err)
}

// TestApproveRequest tests approving a pending request
func TestApproveRequest(t *testing.T) {
	repo := newMockRepository()
	fileHandler := &mockFileHandler{}
	emailService := &mockEmailService{}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	emailConfig := EmailConfig{
		VillageName:  "Test Village",
		SupportEmail: "support@test.com",
		WebsiteURL:   "http://test.com",
	}
	service := NewServiceWithEmail(repo, fileHandler, emailService, clk, jwtSecret, emailConfig, &mockCategoryLookup{})

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

// TestApproveRequestNotFound tests approving a non-existent request
func TestApproveRequestNotFound(t *testing.T) {
	repo := newMockRepository()
	fileHandler := &mockFileHandler{}
	emailService := &mockEmailService{}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	emailConfig := EmailConfig{
		VillageName:  "Test Village",
		SupportEmail: "support@test.com",
		WebsiteURL:   "http://test.com",
	}
	service := NewServiceWithEmail(repo, fileHandler, emailService, clk, jwtSecret, emailConfig, &mockCategoryLookup{})

	_, err := service.ApproveRequest(context.Background(), "non-existent", "admin-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestApproveRequestAlreadyApproved tests that re-approving fails
func TestApproveRequestAlreadyApproved(t *testing.T) {
	repo := newMockRepository()
	fileHandler := &mockFileHandler{}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	service := NewService(repo, fileHandler, clk, jwtSecret, &mockCategoryLookup{})

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
	fileHandler := &mockFileHandler{}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	service := NewService(repo, fileHandler, clk, jwtSecret, &mockCategoryLookup{})

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
	fileHandler := &mockFileHandler{}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	service := NewService(repo, fileHandler, clk, jwtSecret, &mockCategoryLookup{})

	_, err := service.RevokeRequest(context.Background(), "non-existent", "admin-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestRevokeRequestAlreadyRevoked tests that re-revoking fails
func TestRevokeRequestAlreadyRevoked(t *testing.T) {
	repo := newMockRepository()
	fileHandler := &mockFileHandler{}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	service := NewService(repo, fileHandler, clk, jwtSecret, &mockCategoryLookup{})

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
	fileHandler := &mockFileHandler{}
	clk := &mockClock{now: time.Now()}
	jwtSecret := "test-secret-key"
	service := NewService(repo, fileHandler, clk, jwtSecret, &mockCategoryLookup{})

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
