package ppid

import "context"

// SendApprovalEmailInput contains the data needed to send an approval email
type SendApprovalEmailInput struct {
	RequesterEmail      string
	RequesterName       string
	DocumentTitle       string
	DownloadLink        string
	TokenExpirationTime string // ISO 8601 formatted time
	VillageName         string
	SupportEmail        string
	WebsiteURL          string
}

// EmailService defines the interface for sending emails related to PPID requests
// This interface is defined in the usecase layer where it's used, following Go best practices
type EmailService interface {
	// SendApprovalEmail sends an approval notification email to the requester
	// It includes the download link with JWT token and expiration information
	SendApprovalEmail(ctx context.Context, input SendApprovalEmailInput) error
}
