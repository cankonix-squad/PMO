package project

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/governance/workflow"
)

// CreateIssue creates a new issue under a project, enforcing org boundary.
func (s *Service) CreateIssue(ctx context.Context, projectID, orgID, createdBy uuid.UUID, req *CreateIssueRequest) (*Issue, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	if req.TaskID != nil {
		if _, err := s.GetTask(ctx, projectID, *req.TaskID, orgID); err != nil {
			return nil, err
		}
	}

	i := &Issue{
		OrganizationID: orgID,
		ProjectID:      projectID,
		TaskID:         req.TaskID,
		Title:          req.Title,
		Description:    req.Description,
		Status:         "OPEN",
		Severity:       req.Severity,
		Escalation:     req.Escalation,
		ReportedBy:     createdBy,
		AssignedTo:     req.AssignedTo,
		DueDate:        req.DueDate,
		Resolution:     req.Resolution,
	}
	if i.Severity == "" {
		i.Severity = "MEDIUM"
	}
	if i.Escalation == "" {
		i.Escalation = "NONE"
	}
	if err := s.repo.CreateIssue(ctx, i); err != nil {
		return nil, fmt.Errorf("issue: create: %w", err)
	}
	return s.repo.FindIssueByID(ctx, projectID, i.ID, orgID)
}

// GetIssue returns an issue by ID, verifying org and project boundaries.
func (s *Service) GetIssue(ctx context.Context, projectID, issueID, orgID uuid.UUID) (*Issue, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	return s.repo.FindIssueByID(ctx, projectID, issueID, orgID)
}

// ListIssues returns paginated issues for a project.
func (s *Service) ListIssues(ctx context.Context, filter IssueListFilter) ([]Issue, int64, error) {
	if _, err := s.GetByID(ctx, filter.ProjectID, filter.OrganizationID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListIssues(ctx, filter)
}

// UpdateIssue applies partial updates to an issue, enforcing FSM for status.
func (s *Service) UpdateIssue(ctx context.Context, projectID, issueID, orgID uuid.UUID, req *UpdateIssueRequest) (*Issue, error) {
	i, err := s.GetIssue(ctx, projectID, issueID, orgID)
	if err != nil {
		return nil, err
	}
	if req.Title != "" {
		i.Title = req.Title
	}
	if req.Description != "" {
		i.Description = req.Description
	}
	if req.Status != "" {
		if _, err := s.fsm.Transition(workflow.EntityIssue, i.Status, req.Status); err != nil {
			return nil, fmt.Errorf("issue: %w - allowed: %v",
				workflow.ErrTransitionNotAllowed,
				s.fsm.AllowedTransitions(workflow.EntityIssue, i.Status),
			)
		}
		i.Status = req.Status
	}
	if req.Severity != "" {
		i.Severity = req.Severity
	}
	if req.Escalation != "" {
		i.Escalation = req.Escalation
	}
	if req.TaskID != nil {
		if _, err := s.GetTask(ctx, projectID, *req.TaskID, orgID); err != nil {
			return nil, err
		}
		i.TaskID = req.TaskID
	}
	if req.AssignedTo != nil {
		i.AssignedTo = req.AssignedTo
	}
	if req.DueDate != nil {
		i.DueDate = req.DueDate
	}
	if req.Resolution != "" {
		i.Resolution = req.Resolution
	}
	if err := s.repo.UpdateIssue(ctx, i); err != nil {
		return nil, fmt.Errorf("issue: update: %w", err)
	}
	return s.repo.FindIssueByID(ctx, projectID, i.ID, orgID)
}

// DeleteIssue soft-deletes an issue, verifying org boundary.
func (s *Service) DeleteIssue(ctx context.Context, projectID, issueID, orgID uuid.UUID) error {
	i, err := s.GetIssue(ctx, projectID, issueID, orgID)
	if err != nil {
		return err
	}
	return s.repo.DeleteIssue(ctx, projectID, i.ID, orgID)
}
