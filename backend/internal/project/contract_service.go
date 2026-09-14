package project

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/shared/constants"
	"github.com/harmanto-49/cankora/internal/shared/types"
)

// CreateContract creates a project contract, verifying parent project
// ownership, vendor/consultant ownership, and business validation rules.
func (s *Service) CreateContract(ctx context.Context, projectID, orgID, createdBy uuid.UUID, req *CreateContractRequest) (*Contract, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}

	// Vendor must exist in the same org.
	if _, err := s.repo.FindVendorByID(ctx, req.VendorID, orgID); err != nil {
		return nil, ErrVendorNotFound
	}
	var consultantID *uuid.UUID
	if req.ConsultantID != nil {
		if _, err := s.repo.FindVendorByID(ctx, *req.ConsultantID, orgID); err != nil {
			return nil, ErrVendorNotFound
		}
		consultantID = req.ConsultantID
	}

	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status == "" {
		status = constants.ContractStatusDraft
	}
	if !validContractStatuses[status] {
		return nil, ErrInvalidContractStatus
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "IDR"
	}

	value := round2(req.ContractValue)
	if value < 0 {
		return nil, ErrInvalidContractValue
	}
	if err := validateContractDates(req.StartDate, req.EndDate); err != nil {
		return nil, err
	}

	c := &Contract{
		OrganizationID: orgID,
		ProjectID:      projectID,
		ContractNumber: strings.TrimSpace(req.ContractNumber),
		Title:          strings.TrimSpace(req.Title),
		VendorID:       req.VendorID,
		ConsultantID:   consultantID,
		ContractValue:  value,
		Currency:       currency,
		SignedDate:     req.SignedDate,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		Status:         status,
		ScopeOfWork:    req.ScopeOfWork,
		CreatedBy:      createdBy,
	}
	if err := s.repo.CreateContract(ctx, c); err != nil {
		if errors.Is(err, ErrContractNumberTaken) {
			return nil, ErrContractNumberTaken
		}
		return nil, fmt.Errorf("contract: create: %w", err)
	}
	created, err := s.repo.FindContractByID(ctx, projectID, c.ID, orgID)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// ListContracts returns paginated contracts for a project, verifying ownership.
func (s *Service) ListContracts(ctx context.Context, filter ContractListFilter) ([]Contract, int64, error) {
	if _, err := s.GetByID(ctx, filter.ProjectID, filter.OrganizationID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListContracts(ctx, filter)
}

// GetContract returns a contract by ID, verifying parent project ownership.
func (s *Service) GetContract(ctx context.Context, projectID, contractID, orgID uuid.UUID) (*Contract, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	return s.repo.FindContractByID(ctx, projectID, contractID, orgID)
}

// UpdateContract applies updates to a contract, verifying org boundary and
// validating business rules.
func (s *Service) UpdateContract(ctx context.Context, projectID, contractID, orgID uuid.UUID, req *UpdateContractRequest) (*Contract, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	c, err := s.repo.FindContractByID(ctx, projectID, contractID, orgID)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(req.ContractNumber) != "" {
		c.ContractNumber = strings.TrimSpace(req.ContractNumber)
	}
	if strings.TrimSpace(req.Title) != "" {
		c.Title = strings.TrimSpace(req.Title)
	}
	if req.VendorID != nil {
		if _, err := s.repo.FindVendorByID(ctx, *req.VendorID, orgID); err != nil {
			return nil, ErrVendorNotFound
		}
		c.VendorID = *req.VendorID
	}
	if req.ConsultantID != nil {
		if *req.ConsultantID == uuid.Nil {
			c.ConsultantID = nil
		} else {
			if _, err := s.repo.FindVendorByID(ctx, *req.ConsultantID, orgID); err != nil {
				return nil, ErrVendorNotFound
			}
			c.ConsultantID = req.ConsultantID
		}
	}
	if req.ContractValue != nil {
		value := round2(*req.ContractValue)
		if value < 0 {
			return nil, ErrInvalidContractValue
		}
		c.ContractValue = value
	}
	if strings.TrimSpace(req.Currency) != "" {
		c.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	}
	if req.SignedDate != nil {
		c.SignedDate = req.SignedDate
	}
	if req.StartDate != nil {
		c.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		c.EndDate = req.EndDate
	}
	if strings.TrimSpace(req.Status) != "" {
		status := strings.ToUpper(strings.TrimSpace(req.Status))
		if !validContractStatuses[status] {
			return nil, ErrInvalidContractStatus
		}
		c.Status = status
	}
	if req.ScopeOfWork != "" {
		c.ScopeOfWork = req.ScopeOfWork
	}

	if err := validateContractDates(c.StartDate, c.EndDate); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateContract(ctx, c); err != nil {
		if errors.Is(err, ErrContractNumberTaken) {
			return nil, ErrContractNumberTaken
		}
		return nil, fmt.Errorf("contract: update: %w", err)
	}
	updated, err := s.repo.FindContractByID(ctx, projectID, c.ID, orgID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// DeleteContract soft-deletes a contract, verifying org boundary.
func (s *Service) DeleteContract(ctx context.Context, projectID, contractID, orgID uuid.UUID) error {
	if _, err := s.repo.FindContractByID(ctx, projectID, contractID, orgID); err != nil {
		return err
	}
	return s.repo.DeleteContract(ctx, projectID, contractID, orgID)
}

// validateContractDates ensures start_date is not after end_date.
func validateContractDates(start, end *types.FlexTime) error {
	if start == nil || end == nil {
		return nil
	}
	sTime := start.Time
	eTime := end.Time
	if sTime.IsZero() || eTime.IsZero() {
		return nil
	}
	if sTime.After(eTime) {
		return ErrInvalidContractDates
	}
	return nil
}
