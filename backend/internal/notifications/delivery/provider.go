package notification

import "context"

// EmailMessage represents a single outbound email.
type EmailMessage struct {
	To      []string
	CC      []string
	Subject string
	// HTMLBody is the preferred format; TextBody is the fallback.
	HTMLBody string
	TextBody string
}

// Provider is the interface that any email backend must implement.
// This allows swapping SMTP, SendGrid, SES, etc. without touching business logic.
type Provider interface {
	Send(ctx context.Context, msg EmailMessage) error
}

// NoopProvider is a no-op implementation used in development or when SMTP is not configured.
type NoopProvider struct{}

func (n *NoopProvider) Send(_ context.Context, _ EmailMessage) error {
	return nil
}

// Templates holds the named email templates used by the application.
// Each method returns a populated EmailMessage ready to send.
type Templates struct{}

// PasswordReset returns the password reset email.
func (t *Templates) PasswordReset(to, resetURL string) EmailMessage {
	return EmailMessage{
		To:      []string{to},
		Subject: "Reset Your PMO Password",
		HTMLBody: `<p>You requested a password reset for your PMO account.</p>
<p><a href="` + resetURL + `">Click here to reset your password</a></p>
<p>This link expires in 1 hour. If you did not request this, you can safely ignore this email.</p>`,
		TextBody: "Reset your password at: " + resetURL + "\nThis link expires in 1 hour.",
	}
}

// WelcomeUser returns the welcome email sent when a new user is provisioned.
func (t *Templates) WelcomeUser(to, firstName, loginURL string) EmailMessage {
	return EmailMessage{
		To:      []string{to},
		Subject: "Welcome to PMO",
		HTMLBody: `<p>Hi ` + firstName + `,</p>
<p>Your PMO account has been created. You can log in at:</p>
<p><a href="` + loginURL + `">` + loginURL + `</a></p>
<p>You will be prompted to change your password on first login.</p>`,
		TextBody: "Hi " + firstName + ", your PMO account is ready. Login at: " + loginURL,
	}
}
