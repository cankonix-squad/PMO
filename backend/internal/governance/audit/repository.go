package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Repository defines data access for audit logs.
type Repository interface {
	Write(ctx context.Context, log *Log) error
	List(ctx context.Context, filter ListFilter) ([]Log, int64, error)
	GetByID(ctx context.Context, orgID uuid.UUID, id uuid.UUID) (*Log, error)
	Summary(ctx context.Context, orgID uuid.UUID) (*Summary, error)
}

// postgresRepository is the GORM implementation.
type postgresRepository struct {
	db *gorm.DB
}

// NewRepository creates a new audit Repository.
func NewRepository(db *gorm.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Write(ctx context.Context, log *Log) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *postgresRepository) List(ctx context.Context, filter ListFilter) ([]Log, int64, error) {
	query := r.db.WithContext(ctx).Model(&Log{}).
		Where("organization_id = ?", filter.OrganizationID)

	if filter.EntityType != "" {
		query = query.Where("entity_type = ?", filter.EntityType)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.ActorID != nil {
		query = query.Where("actor_id = ?", filter.ActorID)
	}
	if filter.EntityID != "" {
		query = query.Where("entity_id = ?", filter.EntityID)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where(
			"(actor_email ILIKE ? OR action ILIKE ? OR entity_type ILIKE ? OR entity_label ILIKE ?)",
			like, like, like, like,
		)
	}
	if filter.From != nil {
		query = query.Where("created_at >= ?", filter.From)
	}
	if filter.To != nil {
		query = query.Where("created_at <= ?", filter.To)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
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

	var logs []Log
	err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&logs).Error
	return logs, total, err
}

func (r *postgresRepository) GetByID(ctx context.Context, orgID uuid.UUID, id uuid.UUID) (*Log, error) {
	var entry Log
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", orgID, id).
		First(&entry).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrNotFound
	}
	return &entry, err
}

// Summary returns aggregate statistics scoped to the given organization.
func (r *postgresRepository) Summary(ctx context.Context, orgID uuid.UUID) (*Summary, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&Log{}).
		Where("organization_id = ?", orgID).
		Count(&total).Error; err != nil {
		return nil, err
	}

	var uniqueActors int64
	if err := r.db.WithContext(ctx).Model(&Log{}).
		Where("organization_id = ? AND actor_id IS NOT NULL", orgID).
		Distinct("actor_id").Count(&uniqueActors).Error; err != nil {
		return nil, err
	}

	var topActions []ActionCount
	if err := r.db.WithContext(ctx).Model(&Log{}).
		Select("action, count(*) as count").
		Where("organization_id = ?", orgID).
		Group("action").
		Order("count DESC").
		Limit(5).
		Scan(&topActions).Error; err != nil {
		return nil, err
	}

	var topEntities []EntityCount
	if err := r.db.WithContext(ctx).Model(&Log{}).
		Select("entity_type, count(*) as count").
		Where("organization_id = ?", orgID).
		Group("entity_type").
		Order("count DESC").
		Limit(5).
		Scan(&topEntities).Error; err != nil {
		return nil, err
	}

	return &Summary{
		TotalEvents:  total,
		UniqueActors: uniqueActors,
		TopActions:   topActions,
		TopEntities:  topEntities,
	}, nil
}

// Writer is a high-level helper that writes audit logs asynchronously.
// Errors are only logged — audit failures must never block business operations.
type Writer struct {
	repo Repository
	log  *zap.Logger
}

// NewWriter creates a new audit Writer.
func NewWriter(repo Repository, log *zap.Logger) *Writer {
	return &Writer{repo: repo, log: log}
}

// Record writes an audit entry asynchronously.
func (w *Writer) Record(req WriteRequest) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		entry := &Log{
			ID:             uuid.New(),
			OrganizationID: req.OrganizationID,
			ActorID:        req.ActorID,
			ActorEmail:     req.ActorEmail,
			Action:         req.Action,
			EntityType:     req.EntityType,
			EntityID:       req.EntityID,
			EntityLabel:    req.EntityLabel,
			IPAddress:      req.IPAddress,
			UserAgent:      req.UserAgent,
			RequestID:      req.RequestID,
			CreatedAt:      time.Now(),
		}

		if req.OldValues != nil {
			b, err := json.Marshal(req.OldValues)
			if err == nil {
				s := string(b)
				entry.OldValues = &s
			}
		}
		if req.NewValues != nil {
			b, err := json.Marshal(req.NewValues)
			if err == nil {
				s := string(b)
				entry.NewValues = &s
			}
		}

		if err := w.repo.Write(ctx, entry); err != nil {
			w.log.Error("audit: failed to write log",
				zap.String("action", req.Action),
				zap.String("entity_type", req.EntityType),
				zap.String("entity_id", req.EntityID),
				zap.Error(err),
			)
		}
	}()
}
