package user

import (
	"strings"
	"testing"
	"time"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid user",
			user: User{
				ID:             "123e4567-e89b-12d3-a456-426614174000",
				Name:           "John Doe",
				Email:          "john@example.com",
				HashedPassword: "$2a$12$hashedpassword",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid user with profile image",
			user: User{
				ID:              "123e4567-e89b-12d3-a456-426614174000",
				Name:            "Jane Doe",
				Email:           "jane@example.com",
				HashedPassword:  "$2a$12$hashedpassword",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			user: User{
				Name:           "John Doe",
				Email:          "john@example.com",
				HashedPassword: "$2a$12$hashedpassword",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			wantErr: true,
			errMsg:  "user ID is required",
		},
		{
			name: "missing name",
			user: User{
				ID:             "123e4567-e89b-12d3-a456-426614174000",
				Email:          "john@example.com",
				HashedPassword: "$2a$12$hashedpassword",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "empty name after trim",
			user: User{
				ID:             "123e4567-e89b-12d3-a456-426614174000",
				Name:           "   ",
				Email:          "john@example.com",
				HashedPassword: "$2a$12$hashedpassword",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "name too long",
			user: User{
				ID:             "123e4567-e89b-12d3-a456-426614174000",
				Name:           strings.Repeat("a", 256),
				Email:          "john@example.com",
				HashedPassword: "$2a$12$hashedpassword",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			wantErr: true,
			errMsg:  "name must not exceed 255 characters",
		},
		{
			name: "missing email",
			user: User{
				ID:             "123e4567-e89b-12d3-a456-426614174000",
				Name:           "John Doe",
				HashedPassword: "$2a$12$hashedpassword",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name: "invalid email format - no @",
			user: User{
				ID:             "123e4567-e89b-12d3-a456-426614174000",
				Name:           "John Doe",
				Email:          "johnexample.com",
				HashedPassword: "$2a$12$hashedpassword",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			wantErr: true,
			errMsg:  "email must be a valid format",
		},
		{
			name: "invalid email format - no domain",
			user: User{
				ID:             "123e4567-e89b-12d3-a456-426614174000",
				Name:           "John Doe",
				Email:          "john@",
				HashedPassword: "$2a$12$hashedpassword",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			wantErr: true,
			errMsg:  "email must be a valid format",
		},
		{
			name: "email too long",
			user: User{
				ID:             "123e4567-e89b-12d3-a456-426614174000",
				Name:           "John Doe",
				Email:          strings.Repeat("a", 250) + "@test.com",
				HashedPassword: "$2a$12$hashedpassword",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			wantErr: true,
			errMsg:  "email must not exceed 255 characters",
		},
		{
			name: "missing hashed password",
			user: User{
				ID:        "123e4567-e89b-12d3-a456-426614174000",
				Name:      "John Doe",
				Email:     "john@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "hashed password is required",
		},
		{
			name: "missing created_at",
			user: User{
				ID:             "123e4567-e89b-12d3-a456-426614174000",
				Name:           "John Doe",
				Email:          "john@example.com",
				HashedPassword: "$2a$12$hashedpassword",
				UpdatedAt:      time.Now(),
			},
			wantErr: true,
			errMsg:  "created_at timestamp is required",
		},
		{
			name: "missing updated_at",
			user: User{
				ID:             "123e4567-e89b-12d3-a456-426614174000",
				Name:           "John Doe",
				Email:          "john@example.com",
				HashedPassword: "$2a$12$hashedpassword",
				CreatedAt:      time.Now(),
			},
			wantErr: true,
			errMsg:  "updated_at timestamp is required",
		},
	}

		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error but got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
