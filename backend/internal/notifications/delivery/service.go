package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service orchestrates notification creation, delivery, and status updates.
// It writes the record to the DB first (IN_APP always persisted), then
// optionally delivers via the email Provider for EMAIL channel.
type Service struct {
	repo     Repository
	email    Provider
	log      *zap.Logger
	fromAddr string // SMTP from address for email notifications
}

// NewService creates a notification Service.
// emailProvider may be a NoopProvider when SMTP is not configured — it must not be nil.
func NewService(repo Repository, emailProvider Provider, log *zap.Logger, fromAddr string) *Service {
	return &Service{
		repo:     repo,
		email:    emailProvider,
		log:      log,
		fromAddr: fromAddr,
	}
}

// Enqueue persists a notification record (PENDING) and attempts delivery asynchronously.
// The call always returns after the DB write — delivery is fire-and-forget.
// Callers should not block on or retry the returned error for the delivery phase.
func (s *Service) Enqueue(req EnqueueRequest) error {
	if req.Channel == "" {
		req.Channel = ChannelInApp
	}
	if req.Priority == "" {
		req.Priority = PriorityNormal
	}

	n := &Notification{
		ID:              uuid.New(),
		OrganizationID:  req.OrganizationID,
		RecipientUserID: req.RecipientUserID,
		Channel:         req.Channel,
		Status:          StatusPending,
		Priority:        req.Priority,
		Subject:         req.Subject,
		Body:            req.Body,
		SourceType:      req.SourceType,
		SourceID:        req.SourceID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.repo.Create(ctx, n); err != nil {
		s.log.Error("notification: failed to persist",
			zap.String("channel", req.Channel),
			zap.String("subject", req.Subject),
			zap.Error(err),
		)
		return fmt.Errorf("notification: persist: %w", err)
	}

	// IN_APP notifications are stored; no further delivery needed.
	if req.Channel == ChannelInApp {
		now := time.Now()
		_ = s.repo.UpdateStatus(context.Background(), n.ID, StatusSent, "", &now)
		return nil
	}

	// EMAIL: deliver asynchronously so the caller is never blocked.
	go s.deliverEmail(n)
	return nil
}

// EnqueueAndReturn persists a notification and returns the created record.
// Same delivery semantics as Enqueue — DB write is synchronous, email delivery is async.
func (s *Service) EnqueueAndReturn(req EnqueueRequest) (*Notification, error) {
	if req.Channel == "" {
		req.Channel = ChannelInApp
	}
	if req.Priority == "" {
		req.Priority = PriorityNormal
	}

	n := &Notification{
		ID:              uuid.New(),
		OrganizationID:  req.OrganizationID,
		RecipientUserID: req.RecipientUserID,
		Channel:         req.Channel,
		Status:          StatusPending,
		Priority:        req.Priority,
		Subject:         req.Subject,
		Body:            req.Body,
		SourceType:      req.SourceType,
		SourceID:        req.SourceID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.repo.Create(ctx, n); err != nil {
		s.log.Error("notification: failed to persist",
			zap.String("channel", req.Channel),
			zap.String("subject", req.Subject),
			zap.Error(err),
		)
		return nil, fmt.Errorf("notification: persist: %w", err)
	}

	if req.Channel == ChannelInApp {
		now := time.Now()
		_ = s.repo.UpdateStatus(context.Background(), n.ID, StatusSent, "", &now)
		n.Status = StatusSent
		return n, nil
	}

	go s.deliverEmail(n)
	return n, nil
}

func (s *Service) deliverEmail(n *Notification) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	to := ""
	// In a full implementation we'd look up the user email; for UAT foundation
	// we use the body as-is and log. The email provider is a Noop by default.
	msg := EmailMessage{
		To:       []string{to},
		Subject:  n.Subject,
		HTMLBody: fmt.Sprintf("<p>%s</p>", n.Body),
		TextBody: n.Body,
	}

	if err := s.email.Send(ctx, msg); err != nil {
		s.log.Warn("notification: email delivery failed",
			zap.String("id", n.ID.String()),
			zap.Error(err),
		)
		_ = s.repo.UpdateStatus(context.Background(), n.ID, StatusFailed, err.Error(), nil)
		return
	}

	now := time.Now()
	_ = s.repo.UpdateStatus(context.Background(), n.ID, StatusSent, "", &now)
	s.log.Info("notification: email sent", zap.String("id", n.ID.String()))
}

// List returns paginated notifications.
func (s *Service) List(ctx context.Context, filter ListFilter) ([]Notification, int64, error) {
	return s.repo.List(ctx, filter)
}

// GetByID returns a single notification, tenant-scoped.
func (s *Service) GetByID(ctx context.Context, orgID uuid.UUID, id uuid.UUID) (*Notification, error) {
	return s.repo.GetByID(ctx, orgID, id)
}

// MarkRead marks a notification as read.
func (s *Service) MarkRead(ctx context.Context, orgID uuid.UUID, id uuid.UUID) error {
	return s.repo.MarkRead(ctx, orgID, id)
}

// MarkAllRead marks all unread notifications for a recipient as read.
func (s *Service) MarkAllRead(ctx context.Context, orgID uuid.UUID, recipientUserID uuid.UUID) error {
	return s.repo.MarkAllRead(ctx, orgID, recipientUserID)
}

// Retry re-enqueues a FAILED notification for delivery.
func (s *Service) Retry(ctx context.Context, orgID uuid.UUID, id uuid.UUID) error {
	n, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		return err
	}
	if n.Status != StatusFailed {
		return fmt.Errorf("notification: only FAILED notifications can be retried (current: %s)", n.Status)
	}
	// Reset to pending and try again
	_ = s.repo.UpdateStatus(ctx, id, StatusPending, "", nil)
	go s.deliverEmail(n)
	return nil
}

// Summary returns aggregate statistics.
func (s *Service) Summary(ctx context.Context, orgID uuid.UUID, recipientUserID *uuid.UUID) (*Summary, error) {
	return s.repo.Summary(ctx, orgID, recipientUserID)
}
