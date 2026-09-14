package project

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/governance/workflow"
)

// CreateCorrectiveAction creates a new corrective action under a project.
func (s *Service) CreateCorrectiveAction(ctx context.Context, projectID, orgID, createdBy uuid.UUID, req *CreateCorrectiveActionRequest) (*CorrectiveAction, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}

	ca := &CorrectiveAction{
		OrganizationID: orgID,
		ProjectID:      projectID,
		Title:          req.Title,
		Deviation:      req.Deviation,
		RootCause:      req.RootCause,
		Recommendation: req.Recommendation,
		PICUserID:      req.PICUserID,
		TargetDate:     req.TargetDate,
		SourceType:     req.SourceType,
		SourceIssueID:  req.SourceIssueID,
		SourceRiskID:   req.SourceRiskID,
		SourceTaskID:   req.SourceTaskID,
		EvidenceNote:   req.EvidenceNote,
		Status:         "DRAFT",
		CreatedBy:      createdBy,
	}
	if err := s.repo.CreateCorrectiveAction(ctx, ca); err != nil {
		return nil, fmt.Errorf("corrective_action: create: %w", err)
	}
	return s.repo.FindCorrectiveActionByID(ctx, projectID, ca.ID, orgID)
}

// GetCorrectiveAction returns a corrective action by ID, verifying org and project boundaries.
func (s *Service) GetCorrectiveAction(ctx context.Context, projectID, caID, orgID uuid.UUID) (*CorrectiveAction, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	return s.repo.FindCorrectiveActionByID(ctx, projectID, caID, orgID)
}

// ListCorrectiveActions returns paginated corrective actions for a project.
func (s *Service) ListCorrectiveActions(ctx context.Context, filter CorrectiveActionListFilter) ([]CorrectiveAction, int64, error) {
	if _, err := s.GetByID(ctx, filter.ProjectID, filter.OrganizationID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListCorrectiveActions(ctx, filter)
}

// UpdateCorrectiveAction applies partial updates to a corrective action.
// Status changes are validated by the FSM; other field changes are free-form.
func (s *Service) UpdateCorrectiveAction(ctx context.Context, projectID, caID, orgID uuid.UUID, req *UpdateCorrectiveActionRequest) (*CorrectiveAction, error) {
	ca, err := s.GetCorrectiveAction(ctx, projectID, caID, orgID)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		ca.Title = req.Title
	}
	if req.Deviation != "" {
		ca.Deviation = req.Deviation
	}
	if req.RootCause != "" {
		ca.RootCause = req.RootCause
	}
	if req.Recommendation != "" {
		ca.Recommendation = req.Recommendation
	}
	if req.PICUserID != nil {
		ca.PICUserID = req.PICUserID
	}
	if req.TargetDate != nil {
		ca.TargetDate = req.TargetDate
	}
	if req.SourceType != "" {
		ca.SourceType = req.SourceType
	}
	if req.SourceIssueID != nil {
		ca.SourceIssueID = req.SourceIssueID
	}
	if req.SourceRiskID != nil {
		ca.SourceRiskID = req.SourceRiskID
	}
	if req.SourceTaskID != nil {
		ca.SourceTaskID = req.SourceTaskID
	}
	if req.EvidenceNote != "" {
		ca.EvidenceNote = req.EvidenceNote
	}
	// Status updates via UpdateCorrectiveAction are validated by FSM just like
	// the dedicated Transition endpoint, so partial updates can't bypass the FSM.
	if req.Status != "" && req.Status != ca.Status {
		if _, ferr := s.fsm.Transition(workflow.EntityCorrectiveAction, ca.Status, req.Status); ferr != nil {
			return nil, fmt.Errorf("corrective_action: %w — allowed: %v",
				workflow.ErrTransitionNotAllowed,
				s.fsm.AllowedTransitions(workflow.EntityCorrectiveAction, ca.Status),
			)
		}
		ca.Status = req.Status
	}

	if err := s.repo.UpdateCorrectiveAction(ctx, ca); err != nil {
		return nil, fmt.Errorf("corrective_action: update: %w", err)
	}
	return ca, nil
}

// DeleteCorrectiveAction soft-deletes a corrective action.
func (s *Service) DeleteCorrectiveAction(ctx context.Context, projectID, caID, orgID uuid.UUID) error {
	ca, err := s.GetCorrectiveAction(ctx, projectID, caID, orgID)
	if err != nil {
		return err
	}
	return s.repo.DeleteCorrectiveAction(ctx, projectID, ca.ID, orgID)
}

// TransitionCorrectiveAction moves a corrective action to a new status via the FSM.
func (s *Service) TransitionCorrectiveAction(ctx context.Context, projectID, caID, orgID uuid.UUID, req *TransitionRequest) (*CorrectiveAction, error) {
	ca, err := s.GetCorrectiveAction(ctx, projectID, caID, orgID)
	if err != nil {
		return nil, err
	}

	if _, ferr := s.fsm.Transition(workflow.EntityCorrectiveAction, ca.Status, req.ToStatus); ferr != nil {
		return nil, fmt.Errorf("corrective_action: %w — allowed: %v",
			workflow.ErrTransitionNotAllowed,
			s.fsm.AllowedTransitions(workflow.EntityCorrectiveAction, ca.Status),
		)
	}

	ca.Status = req.ToStatus
	if err := s.repo.UpdateCorrectiveAction(ctx, ca); err != nil {
		return nil, fmt.Errorf("corrective_action: transition: %w", err)
	}
	return ca, nil
}
