package kpbu

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/governance/audit"
	"github.com/harmanto-49/cankora/internal/identity/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler serves /api/v1/kpbu endpoints.
type Handler struct {
	svc   *Service
	audit *audit.Writer
	log   *zap.Logger
}

func NewHandler(db *gorm.DB, storageRoot string, maxSize int64, auditWriter *audit.Writer, log *zap.Logger) *Handler {
	return &Handler{svc: NewService(db, log, storageRoot, maxSize), audit: auditWriter, log: log}
}

func claimsFromGin(c *gin.Context) *auth.Claims {
	val, ok := c.Get(string(auth.ContextKeyClaims))
	if !ok {
		return nil
	}
	claims, _ := val.(*auth.Claims)
	return claims
}

func (h *Handler) record(c *gin.Context, claims *auth.Claims, action, entityType string, id uuid.UUID, label string, values interface{}) {
	if h.audit == nil {
		return
	}
	actor := claims.UserID
	h.audit.Record(audit.WriteRequest{
		OrganizationID: claims.OrganizationID,
		ActorID:        &actor,
		ActorEmail:     claims.Email,
		Action:         action,
		EntityType:     entityType,
		EntityID:       id.String(),
		EntityLabel:    label,
		NewValues:      values,
		IPAddress:      c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		RequestID:      c.GetString("X-Request-ID"),
	})
}

// writeErr maps service errors to HTTP responses.
func (h *Handler) writeErr(c *gin.Context, err error, notFoundMsg string) {
	var ve *ValidationError
	var ge *GateError
	switch {
	case errors.Is(err, ErrNotFound):
		response.NotFound(c, notFoundMsg)
	case errors.Is(err, ErrCodeTaken):
		response.Conflict(c, "KPBU project code already in use")
	case errors.Is(err, ErrInvalidStage):
		response.BadRequest(c, "invalid stage")
	case errors.Is(err, ErrInvalidTransition):
		response.BadRequest(c, "stage transition not allowed")
	case errors.As(err, &ve):
		response.BadRequest(c, ve.Msg)
	case errors.As(err, &ge):
		response.Conflict(c, ge.Error())
	default:
		h.log.Error("kpbu: internal error", zap.Error(err))
		response.InternalError(c)
	}
}

func parseID(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		response.BadRequest(c, "invalid "+name)
		return uuid.Nil, false
	}
	return id, true
}

// ListProjects GET /kpbu/projects
func (h *Handler) ListProjects(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	list, err := h.svc.List(claims.OrganizationID, ListFilter{
		Stage: c.Query("stage"), Sector: c.Query("sector"), Search: c.Query("search"),
	})
	if err != nil {
		h.writeErr(c, err, "")
		return
	}
	response.OK(c, "KPBU projects retrieved", list)
}

// GetProject GET /kpbu/projects/:id
func (h *Handler) GetProject(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	d, err := h.svc.Get(claims.OrganizationID, id)
	if err != nil {
		h.writeErr(c, err, "KPBU project not found")
		return
	}
	response.OK(c, "KPBU project retrieved", d)
}

// CreateProject POST /kpbu/projects
func (h *Handler) CreateProject(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	var req ProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.svc.Create(claims.OrganizationID, claims.UserID, req)
	if err != nil {
		h.writeErr(c, err, "")
		return
	}
	h.record(c, claims, "kpbu_project.created", "kpbu_project", p.ID, p.Code+" — "+p.Name, p)
	response.Created(c, "KPBU project created", p)
}

// UpdateProject PUT /kpbu/projects/:id
func (h *Handler) UpdateProject(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req ProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.svc.Update(claims.OrganizationID, id, req)
	if err != nil {
		h.writeErr(c, err, "KPBU project not found")
		return
	}
	h.record(c, claims, "kpbu_project.updated", "kpbu_project", p.ID, p.Code+" — "+p.Name, p)
	response.OK(c, "KPBU project updated", p)
}

// DeleteProject DELETE /kpbu/projects/:id
func (h *Handler) DeleteProject(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	p, err := h.svc.Delete(claims.OrganizationID, id)
	if err != nil {
		h.writeErr(c, err, "KPBU project not found")
		return
	}
	h.record(c, claims, "kpbu_project.deleted", "kpbu_project", p.ID, p.Code+" — "+p.Name, nil)
	response.NoContent(c)
}

// TransitionProject POST /kpbu/projects/:id/transition
func (h *Handler) TransitionProject(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req TransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, from, err := h.svc.Transition(claims.OrganizationID, id, req)
	if err != nil {
		h.writeErr(c, err, "KPBU project not found")
		return
	}
	h.record(c, claims, "kpbu_project.transitioned", "kpbu_project", p.ID, p.Code+" — "+p.Name,
		gin.H{"from_stage": from, "to_stage": p.Stage, "date": req.Date, "notes": req.Notes,
			"override_document_check": req.OverrideDocumentCheck, "override_reason": req.OverrideReason})
	response.OK(c, "KPBU project stage updated", p)
}

// ListStageMilestones GET /kpbu/projects/:id/stage-milestones
func (h *Handler) ListStageMilestones(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.findProject(claims.OrganizationID, id); err != nil {
		h.writeErr(c, err, "KPBU project not found")
		return
	}
	list, err := h.svc.ListStageMilestones(claims.OrganizationID, id)
	if err != nil {
		h.writeErr(c, err, "")
		return
	}
	response.OK(c, "KPBU stage milestones retrieved", list)
}

// UpsertStageMilestone PUT /kpbu/projects/:id/stage-milestones
func (h *Handler) UpsertStageMilestone(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req StageMilestoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	m, err := h.svc.UpsertStageMilestone(claims.OrganizationID, id, req)
	if err != nil {
		h.writeErr(c, err, "KPBU project not found")
		return
	}
	h.record(c, claims, "kpbu_stage_milestone.updated", "kpbu_stage_milestone", m.ID, m.Stage, m)
	response.OK(c, "KPBU stage milestone saved", m)
}

// ListParties GET /kpbu/projects/:id/parties
func (h *Handler) ListParties(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, err := h.svc.findProject(claims.OrganizationID, id); err != nil {
		h.writeErr(c, err, "KPBU project not found")
		return
	}
	list, err := h.svc.ListParties(claims.OrganizationID, id)
	if err != nil {
		h.writeErr(c, err, "")
		return
	}
	response.OK(c, "KPBU parties retrieved", list)
}

// CreateParty POST /kpbu/projects/:id/parties
func (h *Handler) CreateParty(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req PartyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.svc.CreateParty(claims.OrganizationID, id, req)
	if err != nil {
		h.writeErr(c, err, "KPBU project not found")
		return
	}
	h.record(c, claims, "kpbu_party.created", "kpbu_party", p.ID, p.Role+" — "+p.Name, p)
	response.Created(c, "KPBU party created", p)
}

// UpdateParty PUT /kpbu/projects/:id/parties/:partyID
func (h *Handler) UpdateParty(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	partyID, ok := parseID(c, "partyID")
	if !ok {
		return
	}
	var req PartyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.svc.UpdateParty(claims.OrganizationID, id, partyID, req)
	if err != nil {
		h.writeErr(c, err, "KPBU party not found")
		return
	}
	h.record(c, claims, "kpbu_party.updated", "kpbu_party", p.ID, p.Role+" — "+p.Name, p)
	response.OK(c, "KPBU party updated", p)
}

// DeleteParty DELETE /kpbu/projects/:id/parties/:partyID
func (h *Handler) DeleteParty(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	partyID, ok := parseID(c, "partyID")
	if !ok {
		return
	}
	p, err := h.svc.DeleteParty(claims.OrganizationID, id, partyID)
	if err != nil {
		h.writeErr(c, err, "KPBU party not found")
		return
	}
	h.record(c, claims, "kpbu_party.deleted", "kpbu_party", p.ID, p.Role+" — "+p.Name, nil)
	response.NoContent(c)
}
