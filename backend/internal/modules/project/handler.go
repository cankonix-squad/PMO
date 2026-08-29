package project

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/audit"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/core/workflow"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"github.com/harmanto-49/cankora/internal/shared/types"
	"gorm.io/gorm"
)

// Handler wires the project service to HTTP routes.
type Handler struct {
	svc   *Service
	audit *audit.Writer
	db    *gorm.DB // direct DB access for periodic report CRUD (avoids repository interface bloat)
}

// NewHandler creates a new project Handler.
func NewHandler(svc *Service, auditWriter *audit.Writer) *Handler {
	return &Handler{svc: svc, audit: auditWriter}
}

// WithDB sets the gorm.DB reference for direct-DB operations (periodic reports).
// Call this after NewHandler during server wiring.
func (h *Handler) WithDB(db *gorm.DB) *Handler {
	h.db = db
	return h
}

// recordIssueAudit writes an asynchronous audit entry for issue lifecycle events.
func (h *Handler) recordIssueAudit(c *gin.Context, action string, issue *Issue) {
	if h.audit == nil || issue == nil {
		return
	}
	claims := claimsFromGin(c)
	actorID := uuid.Nil
	if claims != nil {
		actorID = claims.UserID
	}
	h.audit.Record(audit.WriteRequest{
		OrganizationID: issue.OrganizationID,
		ActorID:        &actorID,
		ActorEmail:     actorEmailFromClaims(claims),
		Action:         action,
		EntityType:     "issue",
		EntityID:       issue.ID.String(),
		EntityLabel:    issue.Title,
		IPAddress:      c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		RequestID:      c.GetString("X-Request-ID"),
	})
}

// recordRiskAudit writes an asynchronous audit entry for risk lifecycle events.
func (h *Handler) recordRiskAudit(c *gin.Context, action string, risk *Risk) {
	if h.audit == nil || risk == nil {
		return
	}
	claims := claimsFromGin(c)
	actorID := uuid.Nil
	if claims != nil {
		actorID = claims.UserID
	}
	h.audit.Record(audit.WriteRequest{
		OrganizationID: risk.OrganizationID,
		ActorID:        &actorID,
		ActorEmail:     actorEmailFromClaims(claims),
		Action:         action,
		EntityType:     "risk",
		EntityID:       risk.ID.String(),
		EntityLabel:    risk.Title,
		IPAddress:      c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		RequestID:      c.GetString("X-Request-ID"),
	})
}

// recordBudgetAudit writes an asynchronous audit entry for budget lifecycle
// events. Budget lines carry no organization_id directly, so the org is taken
// from the authenticated claims (parent project ownership is already enforced).
func (h *Handler) recordBudgetAudit(c *gin.Context, action string, line *BudgetLine) {
	if h.audit == nil || line == nil {
		return
	}
	claims := claimsFromGin(c)
	if claims == nil {
		return
	}
	actorID := claims.UserID
	h.audit.Record(audit.WriteRequest{
		OrganizationID: claims.OrganizationID,
		ActorID:        &actorID,
		ActorEmail:     actorEmailFromClaims(claims),
		Action:         action,
		EntityType:     "budget",
		EntityID:       line.ID.String(),
		EntityLabel:    line.Category,
		IPAddress:      c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		RequestID:      c.GetString("X-Request-ID"),
	})
}

// recordVendorAudit writes an asynchronous audit entry for vendor lifecycle
// events.
func (h *Handler) recordVendorAudit(c *gin.Context, action string, vendor *Vendor) {
	if h.audit == nil || vendor == nil {
		return
	}
	claims := claimsFromGin(c)
	actorID := uuid.Nil
	if claims != nil {
		actorID = claims.UserID
	}
	h.audit.Record(audit.WriteRequest{
		OrganizationID: vendor.OrganizationID,
		ActorID:        &actorID,
		ActorEmail:     actorEmailFromClaims(claims),
		Action:         action,
		EntityType:     "vendor",
		EntityID:       vendor.ID.String(),
		EntityLabel:    vendor.Name,
		IPAddress:      c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		RequestID:      c.GetString("X-Request-ID"),
	})
}

// recordContractAudit writes an asynchronous audit entry for contract
// lifecycle events.
func (h *Handler) recordContractAudit(c *gin.Context, action string, contract *Contract) {
	if h.audit == nil || contract == nil {
		return
	}
	claims := claimsFromGin(c)
	actorID := uuid.Nil
	if claims != nil {
		actorID = claims.UserID
	}
	label := contract.Title
	if label == "" {
		label = contract.ContractNumber
	}
	h.audit.Record(audit.WriteRequest{
		OrganizationID: contract.OrganizationID,
		ActorID:        &actorID,
		ActorEmail:     actorEmailFromClaims(claims),
		Action:         action,
		EntityType:     "contract",
		EntityID:       contract.ID.String(),
		EntityLabel:    label,
		IPAddress:      c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		RequestID:      c.GetString("X-Request-ID"),
	})
}

func actorEmailFromClaims(claims *auth.Claims) string {
	if claims == nil {
		return ""
	}
	return claims.Email
}

// RegisterRoutes mounts project routes onto a router group.
// The group should already have AuthRequired applied.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.List)
	rg.POST("", h.Create)
	rg.GET("/:id", h.GetByID)
	rg.PUT("/:id", h.Update)
	rg.DELETE("/:id", h.Delete)
	rg.POST("/:id/transition", h.Transition)
	rg.GET("/:id/progress-history", h.GetProgressHistory)
}

// List godoc
// GET /api/v1/projects
func (h *Handler) List(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	page, pageSize := parsePagination(c)
	filter := ProjectListFilter{
		OrganizationID: claims.OrganizationID,
		Status:         c.Query("status"),
		Priority:       c.Query("priority"),
		Search:         c.Query("search"),
		Page:           page,
		PageSize:       pageSize,
	}

	projects, total, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c)
		return
	}

	meta := types.NewPaginationMeta(page, pageSize, total)
	response.OKPaginated(c, "OK", projects, meta)
}

// Create godoc
// POST /api/v1/projects
func (h *Handler) Create(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	p, err := h.svc.Create(c.Request.Context(), &req, claims.OrganizationID, claims.UserID)
	if err != nil {
		if errors.Is(err, ErrCodeTaken) {
			response.Conflict(c, "Project code already in use")
			return
		}
		response.InternalError(c)
		return
	}

	response.Created(c, "Project created", p)
}

// GetByID godoc
// GET /api/v1/projects/:id
func (h *Handler) GetByID(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	id, err := parseUUID(c, "id")
	if err != nil {
		response.BadRequest(c, "Invalid project ID")
		return
	}

	p, err := h.svc.GetByID(c.Request.Context(), id, claims.OrganizationID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "Project not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "OK", p)
}

// Update godoc
// PUT /api/v1/projects/:id
func (h *Handler) Update(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	id, err := parseUUID(c, "id")
	if err != nil {
		response.BadRequest(c, "Invalid project ID")
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	p, err := h.svc.Update(c.Request.Context(), id, claims.OrganizationID, &req)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "Project not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "Project updated", p)
}

// Transition godoc
// POST /api/v1/projects/:id/transition
func (h *Handler) Transition(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	id, err := parseUUID(c, "id")
	if err != nil {
		response.BadRequest(c, "Invalid project ID")
		return
	}

	var req TransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	p, err := h.svc.Transition(c.Request.Context(), id, claims.OrganizationID, &req)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "Project not found")
			return
		}
		if errors.Is(err, workflow.ErrTransitionNotAllowed) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "Status updated", p)
}

// Delete godoc
// DELETE /api/v1/projects/:id
func (h *Handler) Delete(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	id, err := parseUUID(c, "id")
	if err != nil {
		response.BadRequest(c, "Invalid project ID")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id, claims.OrganizationID); err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			response.NotFound(c, "Project not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.NoContent(c)
}

// GetProgressHistory godoc
// GET /api/v1/projects/:id/progress-history
func (h *Handler) GetProgressHistory(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	id, err := parseUUID(c, "id")
	if err != nil {
		response.BadRequest(c, "Invalid project ID")
		return
	}

	history, err := h.svc.GetProgressHistory(c.Request.Context(), id, claims.OrganizationID)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "OK", history)
}

// --- helpers ---

func claimsFromGin(c *gin.Context) *auth.Claims {
	v, exists := c.Get(string(auth.ContextKeyClaims))
	if !exists {
		return nil
	}
	claims, _ := v.(*auth.Claims)
	return claims
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	return uuid.Parse(c.Param(param))
}

func parsePagination(c *gin.Context) (int, int) {
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if v := atoi(p); v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v := atoi(ps); v > 0 && v <= 100 {
			pageSize = v
		}
	}
	return page, pageSize
}

func atoi(s string) int {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}
