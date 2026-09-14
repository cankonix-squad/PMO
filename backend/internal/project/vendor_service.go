package project

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// CreateVendor creates a new vendor/consultant record, scoped to an
// organization.
func (s *Service) CreateVendor(ctx context.Context, orgID, createdBy uuid.UUID, req *CreateVendorRequest) (*Vendor, error) {
	vType := strings.ToUpper(strings.TrimSpace(req.Type))
	if !validVendorTypes[vType] {
		return nil, ErrInvalidVendorType
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	v := &Vendor{
		OrganizationID: orgID,
		Name:           strings.TrimSpace(req.Name),
		Type:           vType,
		LegalName:      strings.TrimSpace(req.LegalName),
		TaxID:          strings.TrimSpace(req.TaxID),
		ContactPerson:  strings.TrimSpace(req.ContactPerson),
		Email:          strings.TrimSpace(req.Email),
		Phone:          strings.TrimSpace(req.Phone),
		Address:        req.Address,
		IsActive:       isActive,
		CreatedBy:      createdBy,
	}
	if err := s.repo.CreateVendor(ctx, v); err != nil {
		return nil, fmt.Errorf("vendor: create: %w", err)
	}
	created, err := s.repo.FindVendorByID(ctx, v.ID, orgID)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// ListVendors returns paginated vendors for an organization.
func (s *Service) ListVendors(ctx context.Context, filter VendorListFilter) ([]Vendor, int64, error) {
	return s.repo.ListVendors(ctx, filter)
}

// GetVendor returns a vendor by ID, verifying org boundary.
func (s *Service) GetVendor(ctx context.Context, id, orgID uuid.UUID) (*Vendor, error) {
	return s.repo.FindVendorByID(ctx, id, orgID)
}

// UpdateVendor applies updates to a vendor, verifying org boundary and
// validating the type enum if provided.
func (s *Service) UpdateVendor(ctx context.Context, id, orgID uuid.UUID, req *UpdateVendorRequest) (*Vendor, error) {
	v, err := s.repo.FindVendorByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(req.Name) != "" {
		v.Name = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.Type) != "" {
		vType := strings.ToUpper(strings.TrimSpace(req.Type))
		if !validVendorTypes[vType] {
			return nil, ErrInvalidVendorType
		}
		v.Type = vType
	}
	if req.LegalName != "" {
		v.LegalName = strings.TrimSpace(req.LegalName)
	}
	if req.TaxID != "" {
		v.TaxID = strings.TrimSpace(req.TaxID)
	}
	if req.ContactPerson != "" {
		v.ContactPerson = strings.TrimSpace(req.ContactPerson)
	}
	if req.Email != "" {
		v.Email = strings.TrimSpace(req.Email)
	}
	if req.Phone != "" {
		v.Phone = strings.TrimSpace(req.Phone)
	}
	if req.Address != "" {
		v.Address = req.Address
	}
	if req.IsActive != nil {
		v.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateVendor(ctx, v); err != nil {
		return nil, fmt.Errorf("vendor: update: %w", err)
	}
	updated, err := s.repo.FindVendorByID(ctx, v.ID, orgID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// DeleteVendor soft-deletes a vendor only if it is not referenced by any
// active contract.
func (s *Service) DeleteVendor(ctx context.Context, id, orgID uuid.UUID) error {
	// Check for active contracts referencing this vendor (as vendor or
	// consultant). Referenced vendors cannot be removed.
	var count int64
	if err := s.repo.countContractsForVendor(ctx, id, orgID, &count); err != nil {
		return err
	}
	if count > 0 {
		return ErrVendorInUse
	}
	return s.repo.DeleteVendor(ctx, id, orgID)
}
