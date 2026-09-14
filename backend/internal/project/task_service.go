package project

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/governance/workflow"
)

// CreateTask creates a new task under a project, enforcing org boundary.
func (s *Service) CreateTask(ctx context.Context, projectID, orgID, createdBy uuid.UUID, req *CreateTaskRequest) (*Task, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	if req.ParentID != nil {
		parent, err := s.GetTask(ctx, projectID, *req.ParentID, orgID)
		if err != nil {
			return nil, err
		}
		if parent.ParentID != nil {
			return nil, fmt.Errorf("task: parent task cannot be nested more than one level")
		}
	}
	if req.MilestoneID != nil {
		if _, err := s.GetMilestone(ctx, projectID, *req.MilestoneID, orgID); err != nil {
			return nil, err
		}
	}

	t := &Task{
		OrganizationID: orgID,
		ProjectID:      projectID,
		MilestoneID:    req.MilestoneID,
		ParentID:       req.ParentID,
		WBSCode:        req.WBSCode,
		Title:          req.Title,
		Description:    req.Description,
		Priority:       req.Priority,
		Type:           req.Type,
		StartDate:      req.StartDate,
		DueDate:        req.DueDate,
		EstHours:       req.EstHours,
		CreatedBy:      createdBy,
	}
	if t.Priority == "" {
		t.Priority = "MEDIUM"
	}
	if t.Type == "" {
		t.Type = "TASK"
	}
	if err := s.repo.CreateTask(ctx, t); err != nil {
		return nil, fmt.Errorf("task: create: %w", err)
	}
	return s.repo.FindTaskByID(ctx, projectID, t.ID, orgID)
}

// GetTask returns a task by ID, verifying org boundary.
func (s *Service) GetTask(ctx context.Context, projectID, taskID, orgID uuid.UUID) (*Task, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	return s.repo.FindTaskByID(ctx, projectID, taskID, orgID)
}

// ListTasks returns paginated tasks for a project.
func (s *Service) ListTasks(ctx context.Context, filter TaskListFilter) ([]Task, int64, error) {
	if _, err := s.GetByID(ctx, filter.ProjectID, filter.OrganizationID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListTasks(ctx, filter)
}

// UpdateTask applies partial updates to a task.
func (s *Service) UpdateTask(ctx context.Context, projectID, taskID, orgID uuid.UUID, req *UpdateTaskRequest) (*Task, error) {
	t, err := s.GetTask(ctx, projectID, taskID, orgID)
	if err != nil {
		return nil, err
	}
	if req.Title != "" {
		t.Title = req.Title
	}
	if req.Description != "" {
		t.Description = req.Description
	}
	if req.Status != "" {
		if _, err := s.fsm.Transition(workflow.EntityTask, t.Status, req.Status); err != nil {
			return nil, fmt.Errorf("task: %w - allowed: %v",
				workflow.ErrTransitionNotAllowed,
				s.fsm.AllowedTransitions(workflow.EntityTask, t.Status),
			)
		}
		t.Status = req.Status
	}
	if req.Priority != "" {
		t.Priority = req.Priority
	}
	if req.Type != "" {
		t.Type = req.Type
	}
	if req.WBSCode != "" {
		t.WBSCode = req.WBSCode
	}
	if req.MilestoneID != nil {
		if _, err := s.GetMilestone(ctx, projectID, *req.MilestoneID, orgID); err != nil {
			return nil, err
		}
		t.MilestoneID = req.MilestoneID
	}
	if req.StartDate != nil {
		t.StartDate = req.StartDate
	}
	if req.DueDate != nil {
		t.DueDate = req.DueDate
	}
	if req.EstHours > 0 {
		t.EstHours = req.EstHours
	}
	if req.ActualHours != nil {
		t.ActualHours = *req.ActualHours
	}
	if req.ProgressPct != nil {
		t.ProgressPct = *req.ProgressPct
	}
	if err := s.repo.UpdateTask(ctx, t); err != nil {
		return nil, fmt.Errorf("task: update: %w", err)
	}
	return s.repo.FindTaskByID(ctx, projectID, t.ID, orgID)
}

// DeleteTask soft-deletes a task, verifying org boundary.
func (s *Service) DeleteTask(ctx context.Context, projectID, taskID, orgID uuid.UUID) error {
	t, err := s.GetTask(ctx, projectID, taskID, orgID)
	if err != nil {
		return err
	}
	return s.repo.DeleteTask(ctx, projectID, t.ID, orgID)
}
