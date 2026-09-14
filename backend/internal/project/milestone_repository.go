package project

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *postgresRepository) CreateMilestone(ctx context.Context, m *Milestone) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *postgresRepository) FindMilestoneByID(ctx context.Context, projectID, milestoneID, orgID uuid.UUID) (*Milestone, error) {
	var m Milestone
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", milestoneID, projectID, orgID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMilestoneNotFound
	}
	return &m, err
}

func (r *postgresRepository) ListMilestones(ctx context.Context, projectID, orgID uuid.UUID) ([]Milestone, error) {
	var milestones []Milestone
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND organization_id = ?", projectID, orgID).
		Order("due_date ASC NULLS LAST, created_at ASC").
		Find(&milestones).Error
	return milestones, err
}

func (r *postgresRepository) UpdateMilestone(ctx context.Context, m *Milestone) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *postgresRepository) DeleteMilestone(ctx context.Context, projectID, milestoneID, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", milestoneID, projectID, orgID).
		Delete(&Milestone{}).Error
}
