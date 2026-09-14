package project

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
)

// CreateBudget creates a budget line under a project, verifying org boundary.
func (s *Service) CreateBudget(ctx context.Context, projectID, orgID, createdBy uuid.UUID, req *CreateBudgetRequest) (*BudgetLine, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}

	planned := round2(req.Planned)
	if planned < 0 {
		return nil, fmt.Errorf("budget: planned cannot be negative")
	}
	actual := round2(req.Actual)
	if actual < 0 {
		return nil, fmt.Errorf("budget: actual cannot be negative")
	}
	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "IDR"
	}

	b := &ProjectBudget{
		ProjectID:   projectID,
		Category:    req.Category,
		Description: req.Description,
		Planned:     planned,
		Actual:      actual,
		Currency:    currency,
		CreatedBy:   createdBy,
	}
	if err := s.repo.CreateBudget(ctx, b); err != nil {
		return nil, fmt.Errorf("budget: create: %w", err)
	}
	line, err := s.repo.FindBudgetByID(ctx, projectID, b.ID, orgID)
	if err != nil {
		return nil, err
	}
	return buildBudgetLine(line), nil
}

// ListBudgets returns paginated budget lines for a project.
func (s *Service) ListBudgets(ctx context.Context, filter BudgetListFilter) ([]BudgetLine, int64, error) {
	if _, err := s.GetByID(ctx, filter.ProjectID, filter.OrganizationID); err != nil {
		return nil, 0, err
	}
	rows, total, err := s.repo.ListBudgets(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	lines := make([]BudgetLine, 0, len(rows))
	for _, r := range rows {
		lines = append(lines, *buildBudgetLine(&r))
	}
	return lines, total, nil
}

// GetBudget returns a budget line by ID, verifying parent project ownership.
func (s *Service) GetBudget(ctx context.Context, projectID, budgetID, orgID uuid.UUID) (*BudgetLine, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	line, err := s.repo.FindBudgetByID(ctx, projectID, budgetID, orgID)
	if err != nil {
		return nil, err
	}
	return buildBudgetLine(line), nil
}

// UpdateBudget applies partial updates to a budget line, verifying org boundary.
func (s *Service) UpdateBudget(ctx context.Context, projectID, budgetID, orgID uuid.UUID, req *UpdateBudgetRequest) (*BudgetLine, error) {
	if _, err := s.GetByID(ctx, projectID, orgID); err != nil {
		return nil, err
	}
	line, err := s.repo.FindBudgetByID(ctx, projectID, budgetID, orgID)
	if err != nil {
		return nil, err
	}

	if req.Category != "" {
		line.Category = req.Category
	}
	if req.Description != "" {
		line.Description = req.Description
	}
	if req.Planned != nil {
		if *req.Planned < 0 {
			return nil, fmt.Errorf("budget: planned cannot be negative")
		}
		line.Planned = round2(*req.Planned)
	}
	if req.Actual != nil {
		if *req.Actual < 0 {
			return nil, fmt.Errorf("budget: actual cannot be negative")
		}
		line.Actual = round2(*req.Actual)
	}
	if strings.TrimSpace(req.Currency) != "" {
		line.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	}

	if err := s.repo.UpdateBudget(ctx, line); err != nil {
		return nil, fmt.Errorf("budget: update: %w", err)
	}
	updated, err := s.repo.FindBudgetByID(ctx, projectID, line.ID, orgID)
	if err != nil {
		return nil, err
	}
	return buildBudgetLine(updated), nil
}

// DeleteBudget soft-deletes a budget line, verifying org boundary.
func (s *Service) DeleteBudget(ctx context.Context, projectID, budgetID, orgID uuid.UUID) error {
	line, err := s.repo.FindBudgetByID(ctx, projectID, budgetID, orgID)
	if err != nil {
		return err
	}
	return s.repo.DeleteBudget(ctx, projectID, line.ID, orgID)
}

// round2 rounds a money amount to 2 decimal places.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// buildBudgetLine computes derived fields (variance, usage_pct, status) for a
// budget line. Variance = planned - actual (positive means remaining budget).
// UsagePct = actual / planned * 100, and is 0 when planned is 0 to avoid
// division by zero. Status thresholds: <80% NORMAL, >=80% WATCH, >=90% RISK,
// >=100% OVERRUN.
func buildBudgetLine(b *ProjectBudget) *BudgetLine {
	planned := b.Planned
	actual := b.Actual
	variance := round2(planned - actual)

	var usagePct float64
	if planned > 0 {
		usagePct = math.Round((actual/planned)*10000) / 100
	}

	status := BudgetStatusNormal
	switch {
	case usagePct >= 100:
		status = BudgetStatusOverrun
	case usagePct >= 90:
		status = BudgetStatusRisk
	case usagePct >= 80:
		status = BudgetStatusWatch
	}

	return &BudgetLine{
		ID:          b.ID,
		ProjectID:   b.ProjectID,
		Category:    b.Category,
		Description: b.Description,
		Planned:     b.Planned,
		Actual:      b.Actual,
		Currency:    b.Currency,
		Variance:    variance,
		UsagePct:    usagePct,
		Status:      status,
		CreatedBy:   b.CreatedBy,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}
