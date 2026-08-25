package auth

import "context"

// EmailSender is the subset of the auth email service the auth.Service
// needs. Defined as an interface here (consumer side) so the usecase layer
// doesn't import the email implementation.
//
// Implemented by email.AuthEmailService.
type EmailSender interface {
	SendPasswordResetEmail(ctx context.Context, recipientEmail, recipientName, resetLink string) error
}