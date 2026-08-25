package pagination

import (
	"testing"
)

func TestParams_Validate(t *testing.T) {
	tests := []struct {
		name      string
		params    Params
		wantPage  int
		wantLimit int
		wantErr   bool
	}{
		{
			name:      "valid params",
			params:    Params{Page: 2, Limit: 20},
			wantPage:  2,
			wantLimit: 20,
			wantErr:   false,
		},
		{
			name:      "default page when zero",
			params:    Params{Page: 0, Limit: 20},
			wantPage:  DefaultPage,
			wantLimit: 20,
			wantErr:   false,
		},
		{
			name:      "default page when negative",
			params:    Params{Page: -1, Limit: 20},
			wantPage:  DefaultPage,
			wantLimit: 20,
			wantErr:   false,
		},
		{
			name:      "default limit when zero",
			params:    Params{Page: 1, Limit: 0},
			wantPage:  1,
			wantLimit: DefaultLimit,
			wantErr:   false,
		},
		{
			name:      "default limit when negative",
			params:    Params{Page: 1, Limit: -1},
			wantPage:  1,
			wantLimit: DefaultLimit,
			wantErr:   false,
		},
		{
			name:      "both defaults",
			params:    Params{Page: 0, Limit: 0},
			wantPage:  DefaultPage,
			wantLimit: DefaultLimit,
			wantErr:   false,
		},
		{
			name:      "max limit allowed",
			params:    Params{Page: 1, Limit: MaxLimit},
			wantPage:  1,
			wantLimit: MaxLimit,
			wantErr:   false,
		},
		{
			name:      "limit exceeds max - should error",
			params:    Params{Page: 1, Limit: MaxLimit + 1},
			wantPage:  1,
			wantLimit: MaxLimit + 1,
			wantErr:   true,
		},
		{
			name:      "limit far exceeds max - should error",
			params:    Params{Page: 1, Limit: 500},
			wantPage:  1,
			wantLimit: 500,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.params.Page != tt.wantPage {
					t.Errorf("Validate() page = %v, want %v", tt.params.Page, tt.wantPage)
				}
				if tt.params.Limit != tt.wantLimit {
					t.Errorf("Validate() limit = %v, want %v", tt.params.Limit, tt.wantLimit)
				}
			}
		})
	}
}

func TestParams_Offset(t *testing.T) {
	tests := []struct {
		name       string
		params     Params
		wantOffset int
	}{
		{
			name:       "first page",
			params:     Params{Page: 1, Limit: 10},
			wantOffset: 0,
		},
		{
			name:       "second page",
			params:     Params{Page: 2, Limit: 10},
			wantOffset: 10,
		},
		{
			name:       "third page",
			params:     Params{Page: 3, Limit: 10},
			wantOffset: 20,
		},
		{
			name:       "page 5 with limit 20",
			params:     Params{Page: 5, Limit: 20},
			wantOffset: 80,
		},
		{
			name:       "page 10 with limit 50",
			params:     Params{Page: 10, Limit: 50},
			wantOffset: 450,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.params.Offset(); got != tt.wantOffset {
				t.Errorf("Offset() = %v, want %v", got, tt.wantOffset)
			}
		})
	}
}

func TestNewResult(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		limit          int
		total          int
		wantTotalPages int
	}{
		{
			name:           "exact division",
			page:           1,
			limit:          10,
			total:          100,
			wantTotalPages: 10,
		},
		{
			name:           "with remainder",
			page:           1,
			limit:          10,
			total:          95,
			wantTotalPages: 10,
		},
		{
			name:           "single page",
			page:           1,
			limit:          10,
			total:          5,
			wantTotalPages: 1,
		},
		{
			name:           "empty result",
			page:           1,
			limit:          10,
			total:          0,
			wantTotalPages: 0,
		},
		{
			name:           "large dataset",
			page:           5,
			limit:          50,
			total:          1234,
			wantTotalPages: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewResult(tt.page, tt.limit, tt.total)

			if result.Page != tt.page {
				t.Errorf("NewResult() page = %v, want %v", result.Page, tt.page)
			}
			if result.Limit != tt.limit {
				t.Errorf("NewResult() limit = %v, want %v", result.Limit, tt.limit)
			}
			if result.Total != tt.total {
				t.Errorf("NewResult() total = %v, want %v", result.Total, tt.total)
			}
			if result.TotalPages != tt.wantTotalPages {
				t.Errorf("NewResult() totalPages = %v, want %v", result.TotalPages, tt.wantTotalPages)
			}
		})
	}
}

func TestPaginate(t *testing.T) {
	tests := []struct {
		name        string
		page        int
		limit       int
		wantOffset  int
		wantLimit   int
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid pagination",
			page:       2,
			limit:      20,
			wantOffset: 20,
			wantLimit:  20,
			wantErr:    false,
		},
		{
			name:       "first page",
			page:       1,
			limit:      10,
			wantOffset: 0,
			wantLimit:  10,
			wantErr:    false,
		},
		{
			name:       "use defaults",
			page:       0,
			limit:      0,
			wantOffset: 0,
			wantLimit:  DefaultLimit,
			wantErr:    false,
		},
		{
			name:       "max limit allowed",
			page:       1,
			limit:      MaxLimit,
			wantOffset: 0,
			wantLimit:  MaxLimit,
			wantErr:    false,
		},
		{
			name:        "limit exceeds max",
			page:        1,
			limit:       101,
			wantOffset:  0,
			wantLimit:   0,
			wantErr:     true,
			errContains: "cannot exceed",
		},
		{
			name:       "large page number",
			page:       100,
			limit:      50,
			wantOffset: 4950,
			wantLimit:  50,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset, limit, err := Paginate(tt.page, tt.limit)

			if (err != nil) != tt.wantErr {
				t.Errorf("Paginate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("Paginate() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if !tt.wantErr {
				if offset != tt.wantOffset {
					t.Errorf("Paginate() offset = %v, want %v", offset, tt.wantOffset)
				}
				if limit != tt.wantLimit {
					t.Errorf("Paginate() limit = %v, want %v", limit, tt.wantLimit)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
