package ppid

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPPIDRequestStatusValidation tests status field validation
func TestPPIDRequestStatusValidation(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{"valid pending", "pending", false},
		{"valid approved", "approved", false},
		{"valid revoked", "revoked", false},
		{"invalid status", "invalid", true},
		{"empty status", "", true},
		{"whitespace status", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequestStatus(tt.status)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestPPIDRequestApprovedValidation tests that approved requests require both ApprovedAt and ApprovedBy
func TestPPIDRequestApprovedValidation(t *testing.T) {
	now := time.Now()
	userID := "user-123"

	tests := []struct {
		name       string
		status     string
		approvedAt *time.Time
		approvedBy *string
		wantErr    bool
	}{
		{
			name:       "approved with both fields",
			status:     "approved",
			approvedAt: &now,
			approvedBy: &userID,
			wantErr:    false,
		},
		{
			name:       "approved without approvedAt",
			status:     "approved",
			approvedAt: nil,
			approvedBy: &userID,
			wantErr:    true,
		},
		{
			name:       "approved without approvedBy",
			status:     "approved",
			approvedAt: &now,
			approvedBy: nil,
			wantErr:    true,
		},
		{
			name:       "pending without approval fields",
			status:     "pending",
			approvedAt: nil,
			approvedBy: nil,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &PPIDRequest{
				ID:             "req-123",
				PPIDId:         "ppid-123",
				RequesterName:  "John Doe",
				RequesterEmail: "john@example.com",
				Status:         tt.status,
				ApprovedAt:     tt.approvedAt,
				ApprovedBy:     tt.approvedBy,
				CreatedAt:      time.Now(),
			}

			err := req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestPPIDRequestRevokedValidation tests that revoked requests require both RevokedAt and RevokedBy
func TestPPIDRequestRevokedValidation(t *testing.T) {
	now := time.Now()
	userID := "user-123"

	tests := []struct {
		name      string
		status    string
		revokedAt *time.Time
		revokedBy *string
		wantErr   bool
	}{
		{
			name:      "revoked with both fields",
			status:    "revoked",
			revokedAt: &now,
			revokedBy: &userID,
			wantErr:   false,
		},
		{
			name:      "revoked without revokedAt",
			status:    "revoked",
			revokedAt: nil,
			revokedBy: &userID,
			wantErr:   true,
		},
		{
			name:      "revoked without revokedBy",
			status:    "revoked",
			revokedAt: &now,
			revokedBy: nil,
			wantErr:   true,
		},
		{
			name:      "pending without revocation fields",
			status:    "pending",
			revokedAt: nil,
			revokedBy: nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &PPIDRequest{
				ID:             "req-123",
				PPIDId:         "ppid-123",
				RequesterName:  "John Doe",
				RequesterEmail: "john@example.com",
				Status:         tt.status,
				RevokedAt:      tt.revokedAt,
				RevokedBy:      tt.revokedBy,
				CreatedAt:      time.Now(),
			}

			err := req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestPPIDRequestHelperMethods tests IsApproved, IsRevoked, and IsPending methods
func TestPPIDRequestHelperMethods(t *testing.T) {
	tests := []struct {
		name         string
		status       string
		wantApproved bool
		wantRevoked  bool
		wantPending  bool
	}{
		{"pending status", "pending", false, false, true},
		{"approved status", "approved", true, false, false},
		{"revoked status", "revoked", false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &PPIDRequest{
				ID:             "req-123",
				PPIDId:         "ppid-123",
				RequesterName:  "John Doe",
				RequesterEmail: "john@example.com",
				Status:         tt.status,
				CreatedAt:      time.Now(),
			}

			assert.Equal(t, tt.wantApproved, req.IsApproved())
			assert.Equal(t, tt.wantRevoked, req.IsRevoked())
			assert.Equal(t, tt.wantPending, req.IsPending())
		})
	}
}

// TestPPIDRequestValidation tests complete PPIDRequest validation
func TestPPIDRequestValidation(t *testing.T) {
	validRequest := &PPIDRequest{
		ID:             "req-123",
		PPIDId:         "ppid-123",
		RequesterName:  "John Doe",
		RequesterEmail: "john@example.com",
		Purpose:        nil,
		Status:         "pending",
		CreatedAt:      time.Now(),
	}

	// Valid request should pass
	err := validRequest.Validate()
	require.NoError(t, err)

	// Missing ID should fail
	invalidReq := *validRequest
	invalidReq.ID = ""
	err = invalidReq.Validate()
	assert.Error(t, err)

	// Missing PPIDId should fail
	invalidReq = *validRequest
	invalidReq.PPIDId = ""
	err = invalidReq.Validate()
	assert.Error(t, err)

	// Missing RequesterName should fail
	invalidReq = *validRequest
	invalidReq.RequesterName = ""
	err = invalidReq.Validate()
	assert.Error(t, err)

	// Missing RequesterEmail should fail
	invalidReq = *validRequest
	invalidReq.RequesterEmail = ""
	err = invalidReq.Validate()
	assert.Error(t, err)

	// Invalid email should fail
	invalidReq = *validRequest
	invalidReq.RequesterEmail = "invalid-email"
	err = invalidReq.Validate()
	assert.Error(t, err)

	// Zero CreatedAt should fail
	invalidReq = *validRequest
	invalidReq.CreatedAt = time.Time{}
	err = invalidReq.Validate()
	assert.Error(t, err)
}

// TestPPIDValidation tests complete PPID validation
func TestPPIDValidation(t *testing.T) {
	validPPID := &PPID{
		ID:              "ppid-123",
		Title:           "Test Document",
		DocumentMediaID: ptr("00000000-0000-0000-0000-000000000001"),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Valid PPID should pass
	err := validPPID.Validate()
	require.NoError(t, err)

	// Missing ID should fail
	invalidPPID := *validPPID
	invalidPPID.ID = ""
	err = invalidPPID.Validate()
	assert.Error(t, err)

	// Missing Title should fail
	invalidPPID = *validPPID
	invalidPPID.Title = ""
	err = invalidPPID.Validate()
	assert.Error(t, err)

	// Zero CreatedAt should fail
	invalidPPID = *validPPID
	invalidPPID.CreatedAt = time.Time{}
	err = invalidPPID.Validate()
	assert.Error(t, err)

	// Zero UpdatedAt should fail
	invalidPPID = *validPPID
	invalidPPID.UpdatedAt = time.Time{}
	err = invalidPPID.Validate()
	assert.Error(t, err)
}

func ptr(s string) *string { return &s }
