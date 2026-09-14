package project

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *postgresRepository) CreateTask(ctx context.Context, t *Task) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *postgresRepository) FindTaskByID(ctx context.Context, projectID, taskID, orgID uuid.UUID) (*Task, error) {
	var t Task
	err := r.db.WithContext(ctx).
		Preload("Subtasks", "organization_id = ? AND project_id = ?", orgID, projectID).
		Preload("Assignments").
		Where("id = ? AND project_id = ? AND organization_id = ?", taskID, projectID, orgID).
		First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTaskNotFound
	}
	return &t, err
}

func (r *postgresRepository) ListTasks(ctx context.Context, filter TaskListFilter) ([]Task, int64, error) {
	q := r.db.WithContext(ctx).Model(&Task{}).
		Where("project_id = ? AND organization_id = ?", filter.ProjectID, filter.OrganizationID).
		Where("parent_id IS NULL") // top-level tasks only

	if filter.MilestoneID != nil {
		q = q.Where("milestone_id = ?", filter.MilestoneID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.AssignedTo != nil {
		q = q.Joins("JOIN task_assignments ON task_assignments.task_id = tasks.id").
			Where("task_assignments.user_id = ?", filter.AssignedTo)
	}
	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		q = q.Where("title ILIKE ?", s)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	var tasks []Task
	err := q.Preload("Subtasks").Preload("Assignments").
		Order("created_at ASC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&tasks).Error
	return tasks, total, err
}

func (r *postgresRepository) UpdateTask(ctx context.Context, t *Task) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *postgresRepository) DeleteTask(ctx context.Context, projectID, taskID, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND organization_id = ?", taskID, projectID, orgID).
		Delete(&Task{}).Error
}
