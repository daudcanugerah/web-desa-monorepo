package pagination

import "fmt"

const (
	// DefaultPage is the default page number when not specified
	DefaultPage = 1

	// DefaultLimit is the default number of items per page
	DefaultLimit = 10

	// MaxLimit is the maximum allowed items per page
	MaxLimit = 100
)

// Params represents pagination parameters
type Params struct {
	Page  int
	Limit int
}

// Result represents pagination metadata
type Result struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Validate validates and normalizes pagination parameters
// Returns error if limit exceeds MaxLimit (Requirement 17.3)
func (p *Params) Validate() error {
	// Set defaults if not specified (Requirement 17.1)
	if p.Page <= 0 {
		p.Page = DefaultPage
	}

	if p.Limit <= 0 {
		p.Limit = DefaultLimit
	}

	// Enforce maximum limit (Requirement 17.3)
	if p.Limit > MaxLimit {
		return fmt.Errorf("limit cannot exceed %d", MaxLimit)
	}

	return nil
}

// Offset calculates the database offset for the current page
func (p *Params) Offset() int {
	return (p.Page - 1) * p.Limit
}

// NewResult creates pagination metadata from params and total count
func NewResult(page, limit, total int) Result {
	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	return Result{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// Paginate is a convenience function that validates params and returns offset
// Returns error if validation fails (Requirement 17.3)
func Paginate(page, limit int) (offset int, validatedLimit int, err error) {
	params := &Params{
		Page:  page,
		Limit: limit,
	}

	if err := params.Validate(); err != nil {
		return 0, 0, err
	}

	return params.Offset(), params.Limit, nil
}
