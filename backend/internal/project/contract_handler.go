package project

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"github.com/harmanto-49/cankora/internal/shared/types"
)

// RegisterContractRoutes mounts contract routes under /projects/:id/contracts.
func (h *Handler) RegisterContractRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.ListContracts)
	rg.POST("", h.CreateContract)
	rg.GET("/:contractID", h.GetContract)
	rg.PUT("/:contractID", h.UpdateContract)
	rg.DELETE("/:contractID", h.DeleteContract)
}

// ListContracts godoc
// GET /api/v1/projects/:id/contracts
func (h *Handler) ListContracts(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	page, pageSize := parsePagination(c)

	var vendorID *uuid.UUID
	if v := c.Query("vendor_id"); v != "" {
		if parsed, perr := uuid.Parse(v); perr == nil {
			vendorID = &parsed
		}
	}

	filter := ContractListFilter{
		OrganizationID: claims.OrganizationID,
		ProjectID:      projectID,
		Status:         c.Query("status"),
		VendorID:       vendorID,
		Search:         c.Query("search"),
		Page:           page,
		PageSize:       pageSize,
	}

	contracts, total, err := h.svc.ListContracts(c.Request.Context(), filter)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OKPaginated(c, "contracts retrieved", contracts, types.NewPaginationMeta(page, pageSize, total))
}

// CreateContract godoc
// POST /api/v1/projects/:id/contracts
func (h *Handler) CreateContract(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	var req CreateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	contract, err := h.svc.CreateContract(c.Request.Context(), projectID, claims.OrganizationID, claims.UserID, &req)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrVendorNotFound) {
			response.NotFound(c, "vendor not found")
			return
		}
		if errors.Is(err, ErrInvalidContractStatus) || errors.Is(err, ErrInvalidContractDates) || errors.Is(err, ErrInvalidContractValue) {
			response.BadRequest(c, err.Error())
			return
		}
		if errors.Is(err, ErrContractNumberTaken) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}

	h.recordContractAudit(c, "contract.created", contract)
	response.Created(c, "contract created", contract)
}

// GetContract godoc
// GET /api/v1/projects/:id/contracts/:contractID
func (h *Handler) GetContract(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	contractID, err := uuid.Parse(c.Param("contractID"))
	if err != nil {
		response.BadRequest(c, "invalid contract id")
		return
	}

	contract, err := h.svc.GetContract(c.Request.Context(), projectID, contractID, claims.OrganizationID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrContractNotFound) {
			response.NotFound(c, "contract not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "ok", contract)
}

// UpdateContract godoc
// PUT /api/v1/projects/:id/contracts/:contractID
func (h *Handler) UpdateContract(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	contractID, err := uuid.Parse(c.Param("contractID"))
	if err != nil {
		response.BadRequest(c, "invalid contract id")
		return
	}

	var req UpdateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	contract, err := h.svc.UpdateContract(c.Request.Context(), projectID, contractID, claims.OrganizationID, &req)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "project not found")
			return
		}
		if errors.Is(err, ErrContractNotFound) {
			response.NotFound(c, "contract not found")
			return
		}
		if errors.Is(err, ErrVendorNotFound) {
			response.NotFound(c, "vendor not found")
			return
		}
		if errors.Is(err, ErrInvalidContractStatus) || errors.Is(err, ErrInvalidContractDates) || errors.Is(err, ErrInvalidContractValue) {
			response.BadRequest(c, err.Error())
			return
		}
		if errors.Is(err, ErrContractNumberTaken) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}

	h.recordContractAudit(c, "contract.updated", contract)
	response.OK(c, "ok", contract)
}

// DeleteContract godoc
// DELETE /api/v1/projects/:id/contracts/:contractID
func (h *Handler) DeleteContract(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	contractID, err := uuid.Parse(c.Param("contractID"))
	if err != nil {
		response.BadRequest(c, "invalid contract id")
		return
	}

	contract, getErr := h.svc.GetContract(c.Request.Context(), projectID, contractID, claims.OrganizationID)
	if getErr != nil {
		if errors.Is(getErr, ErrProjectNotFound) || errors.Is(getErr, ErrContractNotFound) {
			response.NotFound(c, "contract not found")
			return
		}
		response.InternalError(c)
		return
	}

	if err := h.svc.DeleteContract(c.Request.Context(), projectID, contractID, claims.OrganizationID); err != nil {
		if errors.Is(err, ErrContractNotFound) {
			response.NotFound(c, "contract not found")
			return
		}
		response.InternalError(c)
		return
	}

	h.recordContractAudit(c, "contract.deleted", contract)
	response.NoContent(c)
}
