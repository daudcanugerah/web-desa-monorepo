package email

import (
	"context"
	"fmt"
	"time"

	"webdesa/api/config"
	"webdesa/api/interface/http/middleware"
)

// AuthEmailService sends authentication-related emails (currently
// password reset). It reuses the shared SMTPSender so SMTP credentials and
// connection handling live in one place.
type AuthEmailService struct {
	logger       middleware.Logger
	villageName  string
	supportEmail string
	websiteURL   string
	smtp         *SMTPSender
}

// NewAuthEmailService constructs an AuthEmailService. The cfg pointer may
// be nil or unconfigured; IsConfigured() reports whether SMTP is usable.
func NewAuthEmailService(logger middleware.Logger, cfg *EmailConfig, smtpCfg *config.SMTPConfig) *AuthEmailService {
	villageName := "Village Administration"
	supportEmail := "support@village.go.id"
	websiteURL := "http://localhost:8080"

	if cfg != nil {
		if cfg.VillageName != "" {
			villageName = cfg.VillageName
		}
		if cfg.SupportEmail != "" {
			supportEmail = cfg.SupportEmail
		}
		if cfg.WebsiteURL != "" {
			websiteURL = cfg.WebsiteURL
		}
	}

	return &AuthEmailService{
		logger:       logger,
		villageName:  villageName,
		supportEmail: supportEmail,
		websiteURL:   websiteURL,
		smtp:         NewSMTPSender(smtpCfg),
	}
}

// SendPasswordResetEmail sends a password-reset link to the user.
// resetLink is the fully-qualified URL the user clicks to complete the
// reset flow (must include the token query parameter).
//
// Returns an error if SMTP is not configured or the underlying SendMail
// call fails. Callers should surface this error to the operator — the
// silent-failure pattern of always returning success (see bugs.md#1)
// has been fixed.
func (s *AuthEmailService) SendPasswordResetEmail(ctx context.Context, recipientEmail, recipientName, resetLink string) error {
	if !s.smtp.IsConfigured() {
		return fmt.Errorf("SMTP not configured: cannot send password reset email")
	}

	subject := fmt.Sprintf("%s - Password Reset Request", s.villageName)
	body := s.buildPasswordResetEmailBody(recipientName, resetLink)

	if err := s.smtp.Send(recipientEmail, subject, body, "Village Administration"); err != nil {
		s.logger.Error(ctx, "Failed to send password reset email", "error", err.Error())
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	s.logger.Info(ctx, "Password reset email sent",
		"to", recipientEmail,
		"subject", subject,
	)

	return nil
}

func (s *AuthEmailService) buildPasswordResetEmailBody(recipientName, resetLink string) string {
	year := time.Now().Year()
	greeting := "Hello"
	if recipientName != "" {
		greeting = "Hello " + recipientName
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Password Reset</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2c3e50;">Password Reset Request</h2>

        <p>%s,</p>

        <p>We received a request to reset the password for your account at <strong>%s</strong>.</p>

        <p>To reset your password, click the button below. This link will expire in 1 hour.</p>

        <p style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #3498db; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Reset Password</a>
        </p>

        <p>Or copy and paste this URL into your browser:</p>
        <p style="word-break: break-all; background-color: #f8f8f8; padding: 10px; border-left: 3px solid #3498db;">%s</p>

        <p><strong>If you didn't request this,</strong> you can safely ignore this email. Your password will remain unchanged.</p>

        <hr style="border: 1px solid #eee; margin: 30px 0;">

        <p style="color: #7f8c8d; font-size: 12px;">
            Need help? Contact us at <a href="mailto:%s">%s</a><br>
            &copy; %d %s. All rights reserved.
        </p>
    </div>
</body>
</html>
`, greeting, s.villageName, resetLink, resetLink, s.supportEmail, s.supportEmail, year, s.villageName)
}