package notification

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository defines data access for notifications.
type Repository interface {
	Create(ctx context.Context, n *Notification) error
	GetByID(ctx context.Context, orgID uuid.UUID, id uuid.UUID) (*Notification, error)
	List(ctx context.Context, filter ListFilter) ([]Notification, int64, error)
	MarkRead(ctx context.Context, orgID uuid.UUID, id uuid.UUID) error
	MarkAllRead(ctx context.Context, orgID uuid.UUID, recipientUserID uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, errMsg string, sentAt *time.Time) error
	Summary(ctx context.Context, orgID uuid.UUID, recipientUserID *uuid.UUID) (*Summary, error)
}

type postgresRepository struct {
	db *gorm.DB
}

// NewRepository creates a new notification Repository backed by PostgreSQL.
func NewRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, n *Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *postgresRepository) GetByID(ctx context.Context, orgID uuid.UUID, id uuid.UUID) (*Notification, error) {
	var n Notification
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", orgID, id).
		First(&n).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrNotFound
	}
	return &n, err
}

func (r *postgresRepository) List(ctx context.Context, filter ListFilter) ([]Notification, int64, error) {
	q := r.db.WithContext(ctx).Model(&Notification{}).
		Where("organization_id = ?", filter.OrganizationID)

	if filter.RecipientUserID != nil {
		q = q.Where("recipient_user_id = ?", filter.RecipientUserID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Channel != "" {
		q = q.Where("channel = ?", filter.Channel)
	}
	if filter.Priority != "" {
		q = q.Where("priority = ?", filter.Priority)
	}
	if filter.SourceType != "" {
		q = q.Where("source_type = ?", filter.SourceType)
	}
	if filter.UnreadOnly {
		q = q.Where("read_at IS NULL AND status != ?", StatusFailed)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	var items []Notification
	err := q.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&items).Error
	return items, total, err
}

func (r *postgresRepository) MarkRead(ctx context.Context, orgID uuid.UUID, id uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&Notification{}).
		Where("organization_id = ? AND id = ? AND read_at IS NULL", orgID, id).
		Updates(map[string]interface{}{
			"status":  StatusRead,
			"read_at": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		// Already read or not found — check existence
		var count int64
		r.db.WithContext(ctx).Model(&Notification{}).
			Where("organization_id = ? AND id = ?", orgID, id).Count(&count)
		if count == 0 {
			return ErrNotFound
		}
	}
	return nil
}

func (r *postgresRepository) MarkAllRead(ctx context.Context, orgID uuid.UUID, recipientUserID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&Notification{}).
		Where("organization_id = ? AND recipient_user_id = ? AND read_at IS NULL", orgID, recipientUserID).
		Updates(map[string]interface{}{
			"status":  StatusRead,
			"read_at": now,
		}).Error
}

func (r *postgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, errMsg string, sentAt *time.Time) error {
	updates := map[string]interface{}{
		"status":        status,
		"error_message": errMsg,
	}
	if sentAt != nil {
		updates["sent_at"] = sentAt
	}
	return r.db.WithContext(ctx).Model(&Notification{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *postgresRepository) Summary(ctx context.Context, orgID uuid.UUID, recipientUserID *uuid.UUID) (*Summary, error) {
	q := r.db.WithContext(ctx).Model(&Notification{}).Where("organization_id = ?", orgID)
	if recipientUserID != nil {
		q = q.Where("recipient_user_id = ?", recipientUserID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	var unread int64
	uq := r.db.WithContext(ctx).Model(&Notification{}).
		Where("organization_id = ? AND read_at IS NULL AND status != ?", orgID, StatusFailed)
	if recipientUserID != nil {
		uq = uq.Where("recipient_user_id = ?", recipientUserID)
	}
	if err := uq.Count(&unread).Error; err != nil {
		return nil, err
	}

	var pending int64
	pq := r.db.WithContext(ctx).Model(&Notification{}).
		Where("organization_id = ? AND status = ?", orgID, StatusPending)
	if recipientUserID != nil {
		pq = pq.Where("recipient_user_id = ?", recipientUserID)
	}
	if err := pq.Count(&pending).Error; err != nil {
		return nil, err
	}

	var failed int64
	fq := r.db.WithContext(ctx).Model(&Notification{}).
		Where("organization_id = ? AND status = ?", orgID, StatusFailed)
	if recipientUserID != nil {
		fq = fq.Where("recipient_user_id = ?", recipientUserID)
	}
	if err := fq.Count(&failed).Error; err != nil {
		return nil, err
	}

	return &Summary{
		Total:   total,
		Unread:  unread,
		Pending: pending,
		Failed:  failed,
	}, nil
}
