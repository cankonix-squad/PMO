package bim

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
)

// Handler exposes BIM/digital twin integration endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new BIM integration Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// claims extracts JWT claims from gin context.
func (h *Handler) claims(c *gin.Context) (*auth.Claims, bool) {
	v, ok := c.Get(string(auth.ContextKeyClaims))
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return nil, false
	}
	cl, ok := v.(*auth.Claims)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return nil, false
	}
	return cl, true
}

// ---------------------------------------------------------------------------
// GET /api/v1/integrations/bim/models
// ---------------------------------------------------------------------------

// ListModels returns all BIM models for the caller's organization.
func (h *Handler) ListModels(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	models, total, err := h.svc.ListModels(cl.OrganizationID)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "bim models retrieved", gin.H{
		"data": models,
		"meta": gin.H{"total": total},
	})
}

// ---------------------------------------------------------------------------
// POST /api/v1/integrations/bim/models
// ---------------------------------------------------------------------------

// CreateModel registers a new BIM model for the caller's organization.
func (h *Handler) CreateModel(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	var req CreateBIMModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	m, err := h.svc.CreateModel(cl.OrganizationID, cl.UserID, req)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.Created(c, "bim model registered", gin.H{"data": m})
}

// ---------------------------------------------------------------------------
// GET /api/v1/integrations/bim/models/:id
// ---------------------------------------------------------------------------

// GetModel returns a single BIM model by ID.
func (h *Handler) GetModel(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	modelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid model id")
		return
	}

	m, err := h.svc.GetModel(cl.OrganizationID, modelID)
	if err != nil {
		response.InternalError(c)
		return
	}
	if m == nil {
		response.NotFound(c, "bim model not found")
		return
	}

	response.OK(c, "bim model retrieved", gin.H{"data": m})
}

// ---------------------------------------------------------------------------
// PATCH /api/v1/integrations/bim/models/:id
// ---------------------------------------------------------------------------

// UpdateModel patches mutable fields of an existing BIM model.
func (h *Handler) UpdateModel(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	modelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid model id")
		return
	}

	var req UpdateBIMModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	m, err := h.svc.UpdateModel(cl.OrganizationID, cl.UserID, modelID, req)
	if err != nil {
		response.InternalError(c)
		return
	}
	if m == nil {
		response.NotFound(c, "bim model not found")
		return
	}

	response.OK(c, "bim model updated", gin.H{"data": m})
}

// ---------------------------------------------------------------------------
// DELETE /api/v1/integrations/bim/models/:id
// ---------------------------------------------------------------------------

// DeleteModel soft-deletes a BIM model.
func (h *Handler) DeleteModel(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	modelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid model id")
		return
	}

	if err := h.svc.DeleteModel(cl.OrganizationID, cl.UserID, modelID); err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "bim model deleted", nil)
}

// ---------------------------------------------------------------------------
// GET /api/v1/integrations/bim/models/:id/versions
// ---------------------------------------------------------------------------

// ListVersions returns all versions for a BIM model.
func (h *Handler) ListVersions(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	modelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid model id")
		return
	}

	versions, err := h.svc.ListVersions(cl.OrganizationID, modelID)
	if err != nil {
		response.InternalError(c)
		return
	}
	if versions == nil {
		response.NotFound(c, "bim model not found")
		return
	}

	response.OK(c, "bim model versions retrieved", gin.H{
		"data": versions,
		"meta": gin.H{"total": len(versions)},
	})
}

// ---------------------------------------------------------------------------
// POST /api/v1/integrations/bim/models/:id/versions
// ---------------------------------------------------------------------------

// AddVersion appends a new immutable version record to a BIM model.
func (h *Handler) AddVersion(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	modelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid model id")
		return
	}

	var req CreateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	v, err := h.svc.AddVersion(cl.OrganizationID, cl.UserID, modelID, req)
	if err != nil {
		response.InternalError(c)
		return
	}
	if v == nil {
		response.NotFound(c, "bim model not found")
		return
	}

	response.Created(c, "bim model version added", gin.H{"data": v})
}

// ---------------------------------------------------------------------------
// GET /api/v1/integrations/bim/models/:id/mappings
// ---------------------------------------------------------------------------

// ListMappings returns all project mappings for a BIM model.
func (h *Handler) ListMappings(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	modelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid model id")
		return
	}

	mappings, err := h.svc.ListMappings(cl.OrganizationID, modelID)
	if err != nil {
		response.InternalError(c)
		return
	}
	if mappings == nil {
		response.NotFound(c, "bim model not found")
		return
	}

	response.OK(c, "bim project mappings retrieved", gin.H{
		"data": mappings,
		"meta": gin.H{"total": len(mappings)},
	})
}

// ---------------------------------------------------------------------------
// POST /api/v1/integrations/bim/models/:id/mappings
// ---------------------------------------------------------------------------

// LinkProject creates a mapping between a BIM model and a project.
func (h *Handler) LinkProject(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	modelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid model id")
		return
	}

	var req LinkProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	mapping, err := h.svc.LinkProject(cl.OrganizationID, cl.UserID, modelID, req)
	if errors.Is(err, ErrProjectNotFound) {
		response.NotFound(c, "project not found or does not belong to your organization")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}
	if mapping == nil {
		response.NotFound(c, "bim model not found")
		return
	}

	response.Created(c, "project linked to bim model", gin.H{"data": mapping})
}

// ---------------------------------------------------------------------------
// DELETE /api/v1/integrations/bim/models/:id/mappings/:project_id
// ---------------------------------------------------------------------------

// UnlinkProject removes a project mapping from a BIM model.
func (h *Handler) UnlinkProject(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	modelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid model id")
		return
	}

	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	if err := h.svc.UnlinkProject(cl.OrganizationID, cl.UserID, modelID, projectID); err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "project unlinked from bim model", nil)
}
