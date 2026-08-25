package email

import (
	"context"
	"fmt"
	"time"

	"webdesa/api/config"
	"webdesa/api/interface/http/middleware"
	"webdesa/api/usecase/ppid"
)

// PPIDEmailService implements the EmailService interface for PPID-related emails
type PPIDEmailService struct {
	logger       middleware.Logger
	villageName  string
	supportEmail string
	websiteURL   string
	smtp         *SMTPSender
}

// EmailConfig holds email customization from database settings
type EmailConfig struct {
	VillageName  string
	SupportEmail string
	WebsiteURL   string
}

// NewPPIDEmailService creates a new PPID email service
func NewPPIDEmailService(logger middleware.Logger, config *EmailConfig, smtpCfg *config.SMTPConfig) *PPIDEmailService {
	villageName := "Village Administration"
	supportEmail := "support@village.go.id"
	websiteURL := "http://localhost:8080"

	if config != nil {
		if config.VillageName != "" {
			villageName = config.VillageName
		}
		if config.SupportEmail != "" {
			supportEmail = config.SupportEmail
		}
		if config.WebsiteURL != "" {
			websiteURL = config.WebsiteURL
		}
	}

	return &PPIDEmailService{
		logger:       logger,
		villageName:  villageName,
		supportEmail: supportEmail,
		websiteURL:   websiteURL,
		smtp:         NewSMTPSender(smtpCfg),
	}
}

// SendApprovalEmail sends an approval notification email to the requester
// This implementation is synchronous and returns errors properly
func (s *PPIDEmailService) SendApprovalEmail(ctx context.Context, input ppid.SendApprovalEmailInput) error {
	// Use input values or fall back to service config
	villageName := input.VillageName
	supportEmail := input.SupportEmail
	websiteURL := input.WebsiteURL

	if villageName == "" {
		villageName = s.villageName
	}
	if supportEmail == "" {
		supportEmail = s.supportEmail
	}
	if websiteURL == "" {
		websiteURL = s.websiteURL
	}

	// Build email subject
	subject := fmt.Sprintf("%s - Your PPID Document Request Has Been Approved", villageName)

	// Build email body with HTML formatting
	body := s.buildApprovalEmailBody(input, villageName, supportEmail, websiteURL)

	if !s.smtp.IsConfigured() {
		return fmt.Errorf("SMTP not configured: cannot send approval email")
	}

	if err := s.smtp.Send(input.RequesterEmail, subject, body, "Village Administration"); err != nil {
		s.logger.Error(ctx, "Failed to send email via SMTP", "error", err.Error())
		return fmt.Errorf("failed to send approval email: %w", err)
	}

	s.logger.Info(ctx, "Email sent via SMTP",
		"to", input.RequesterEmail,
		"subject", subject,
	)

	return nil
}

// buildApprovalEmailBody constructs the HTML email body for approval notifications
func (s *PPIDEmailService) buildApprovalEmailBody(input ppid.SendApprovalEmailInput, villageName, supportEmail, websiteURL string) string {
	year := time.Now().Year()
	if villageName == "" {
		villageName = s.villageName
	}
	if supportEmail == "" {
		supportEmail = s.supportEmail
	}
	if websiteURL == "" {
		websiteURL = s.websiteURL
	}
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 10px 20px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 5px; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
        .info-box { background-color: #e8f5e9; padding: 15px; border-left: 4px solid #4CAF50; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>PPID Document Request Approved</h1>
        </div>
        
        <div class="content">
            <p>Dear %s,</p>
            
            <p>Good news! Your request for the following PPID document has been approved:</p>
            
            <div class="info-box">
                <strong>Document Title:</strong> %s
            </div>
            
            <p>You can now download your document using the link below. Please note that this link will expire in 10 minutes.</p>
            
            <p style="text-align: center; margin: 30px 0;">
                <a href="%s" class="button" target="_blank">Download Document</a>
            </p>
            
            <div class="info-box">
                <strong>Important:</strong> This download link will expire at %s. After that, you will need to submit a new request.
            </div>
            
            <p><strong>Access Instructions:</strong></p>
            <ul>
                <li>Click the "Download Document" button above to download your file</li>
                <li>The link is valid for 10 minutes from the time this email was sent</li>
                <li>If the link expires, please submit a new request</li>
                <li>For security purposes, this link is unique to your request</li>
            </ul>
            
            <p>If you have any questions or need assistance, please contact: %s</p>
            
            <p>Best regards,<br>%s</p>
        </div>
        
        <div class="footer">
            <p>This is an automated email. Please do not reply to this message.</p>
            <p>&copy; %d %s. All rights reserved.</p>
            <p>Visit us at: %s</p>
        </div>
    </div>
</body>
</html>
`, input.RequesterName, input.DocumentTitle, input.DownloadLink, input.TokenExpirationTime, supportEmail, villageName, year, villageName, websiteURL)
}
