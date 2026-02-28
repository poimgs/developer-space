package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// ResendEmailSender sends emails via the Resend HTTP API.
type ResendEmailSender struct {
	apiKey    string
	fromEmail string
	client    *http.Client
}

func NewResendEmailSender(apiKey, fromEmail string) *ResendEmailSender {
	return &ResendEmailSender{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		client:    &http.Client{},
	}
}

func (s *ResendEmailSender) SendInvitation(ctx context.Context, toEmail, toName, frontendURL string) {
	subject := "You've been invited to the co-working space"
	body := fmt.Sprintf(
		"Hi %s,\n\nYou've been invited to the co-working space app.\n\nVisit %s to log in with your email.\n",
		toName, frontendURL,
	)
	if err := s.send(ctx, toEmail, subject, body); err != nil {
		slog.Warn("failed to send invitation email", "email", toEmail, "error", err)
	}
}

func (s *ResendEmailSender) SendNotificationEmail(ctx context.Context, to, subject, body string) {
	if err := s.send(ctx, to, subject, body); err != nil {
		slog.Warn("failed to send notification email", "to", to, "subject", subject, "error", err)
	}
}

func (s *ResendEmailSender) SendMagicLink(ctx context.Context, toEmail, toName, link string) error {
	subject := "Your login link"
	body := fmt.Sprintf(
		"Hi %s,\n\nClick the link below to log in:\n\n%s\n\nThis link expires in 15 minutes.\n",
		toName, link,
	)
	return s.send(ctx, toEmail, subject, body)
}

type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
}

func (s *ResendEmailSender) send(ctx context.Context, to, subject, text string) error {
	payload := resendRequest{
		From:    s.fromEmail,
		To:      []string{to},
		Subject: subject,
		Text:    text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling email request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating email request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API returned status %d", resp.StatusCode)
	}

	slog.Debug("email sent", "to", to, "subject", subject)
	return nil
}
