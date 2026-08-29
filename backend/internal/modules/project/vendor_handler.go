package project

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"github.com/harmanto-49/cankora/internal/shared/types"
)

// RegisterVendorRoutes mounts vendor routes under /api/v1/vendors.
func (h *Handler) RegisterVendorRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.ListVendors)
	rg.POST("", h.CreateVendor)
	rg.GET("/:vendorID", h.GetVendor)
	rg.PUT("/:vendorID", h.UpdateVendor)
	rg.DELETE("/:vendorID", h.DeleteVendor)
}

// ListVendors godoc
// GET /api/v1/vendors?type=VENDOR|CONSULTANT&search=&is_active=
func (h *Handler) ListVendors(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	page, pageSize := parsePagination(c)

	var isActive *bool
	if v := c.Query("is_active"); v != "" {
		b := v == "true" || v == "1"
		isActive = &b
	}

	filter := VendorListFilter{
		OrganizationID: claims.OrganizationID,
		Type:           c.Query("type"),
		Search:         c.Query("search"),
		IsActive:       isActive,
		Page:           page,
		PageSize:       pageSize,
	}

	vendors, total, err := h.svc.ListVendors(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OKPaginated(c, "vendors retrieved", vendors, types.NewPaginationMeta(page, pageSize, total))
}

// CreateVendor godoc
// POST /api/v1/vendors
func (h *Handler) CreateVendor(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	var req CreateVendorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	vendor, err := h.svc.CreateVendor(c.Request.Context(), claims.OrganizationID, claims.UserID, &req)
	if err != nil {
		if errors.Is(err, ErrInvalidVendorType) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}

	h.recordVendorAudit(c, "vendor.created", vendor)
	response.Created(c, "vendor created", vendor)
}

// GetVendor godoc
// GET /api/v1/vendors/:vendorID
func (h *Handler) GetVendor(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	vendorID, err := uuid.Parse(c.Param("vendorID"))
	if err != nil {
		response.BadRequest(c, "invalid vendor id")
		return
	}

	vendor, err := h.svc.GetVendor(c.Request.Context(), vendorID, claims.OrganizationID)
	if err != nil {
		if errors.Is(err, ErrVendorNotFound) {
			response.NotFound(c, "vendor not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "ok", vendor)
}

// UpdateVendor godoc
// PUT /api/v1/vendors/:vendorID
func (h *Handler) UpdateVendor(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	vendorID, err := uuid.Parse(c.Param("vendorID"))
	if err != nil {
		response.BadRequest(c, "invalid vendor id")
		return
	}

	var req UpdateVendorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	vendor, err := h.svc.UpdateVendor(c.Request.Context(), vendorID, claims.OrganizationID, &req)
	if err != nil {
		if errors.Is(err, ErrVendorNotFound) {
			response.NotFound(c, "vendor not found")
			return
		}
		if errors.Is(err, ErrInvalidVendorType) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}

	h.recordVendorAudit(c, "vendor.updated", vendor)
	response.OK(c, "ok", vendor)
}

// DeleteVendor godoc
// DELETE /api/v1/vendors/:vendorID
func (h *Handler) DeleteVendor(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	vendorID, err := uuid.Parse(c.Param("vendorID"))
	if err != nil {
		response.BadRequest(c, "invalid vendor id")
		return
	}

	vendor, getErr := h.svc.GetVendor(c.Request.Context(), vendorID, claims.OrganizationID)
	if getErr != nil {
		if errors.Is(getErr, ErrVendorNotFound) {
			response.NotFound(c, "vendor not found")
			return
		}
		response.InternalError(c)
		return
	}

	if err := h.svc.DeleteVendor(c.Request.Context(), vendorID, claims.OrganizationID); err != nil {
		if errors.Is(err, ErrVendorNotFound) {
			response.NotFound(c, "vendor not found")
			return
		}
		if errors.Is(err, ErrVendorInUse) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}

	h.recordVendorAudit(c, "vendor.deleted", vendor)
	response.NoContent(c)
}
