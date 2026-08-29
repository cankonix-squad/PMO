package commandcenter

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/audit"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"github.com/harmanto-49/cankora/internal/shared/types"
	"gorm.io/gorm"
	"strings"
	"time"
)

type Escalation struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	ProjectID      *uuid.UUID `json:"project_id,omitempty"`
	SourceType     string     `json:"source_type"`
	SourceID       uuid.UUID  `json:"source_id"`
	Level          string     `json:"level"`
	Reason         string     `json:"reason"`
	Status         string     `json:"status"`
	CreatedBy      uuid.UUID  `json:"created_by"`
	AcknowledgedBy *uuid.UUID `json:"acknowledged_by,omitempty"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (Escalation) TableName() string { return "command_escalations" }

type Decision struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	ProjectID      *uuid.UUID      `json:"project_id,omitempty"`
	EscalationID   *uuid.UUID      `json:"escalation_id,omitempty"`
	Subject        string          `json:"subject"`
	DecisionText   string          `json:"decision_text"`
	OwnerID        *uuid.UUID      `json:"owner_id,omitempty"`
	DueDate        *types.FlexTime `json:"due_date,omitempty"`
	Status         string          `json:"status"`
	DecidedBy      *uuid.UUID      `json:"decided_by,omitempty"`
	DecidedAt      *time.Time      `json:"decided_at,omitempty"`
	CreatedBy      uuid.UUID       `json:"created_by"`
	CreatedAt      time.Time       `json:"created_at"`
}

func (Decision) TableName() string { return "executive_decisions" }

type CreateEscalationRequest struct {
	ProjectID  *uuid.UUID `json:"project_id"`
	SourceType string     `json:"source_type" binding:"required,max=50"`
	SourceID   uuid.UUID  `json:"source_id" binding:"required"`
	Level      string     `json:"level" binding:"required,oneof=PROJECT_MANAGER PROGRAM_MANAGER EXECUTIVE"`
	Reason     string     `json:"reason" binding:"required"`
}
type CreateDecisionRequest struct {
	ProjectID    *uuid.UUID      `json:"project_id"`
	EscalationID *uuid.UUID      `json:"escalation_id"`
	Subject      string          `json:"subject" binding:"required,max=500"`
	DecisionText string          `json:"decision_text" binding:"required"`
	OwnerID      *uuid.UUID      `json:"owner_id"`
	DueDate      *types.FlexTime `json:"due_date"`
}
type StatusRequest struct {
	Status string `json:"status" binding:"required"`
}
type CommandHandler struct {
	db    *gorm.DB
	audit *audit.Writer
}

type commandSource struct {
	table      string
	hasProject bool
}

var commandSources = map[string]commandSource{
	"project":             {table: "projects", hasProject: false},
	"issue":               {table: "issues", hasProject: true},
	"risk":                {table: "risks", hasProject: true},
	"corrective_action":   {table: "corrective_actions", hasProject: true},
	"data_submission":     {table: "data_submissions", hasProject: true},
	"health_snapshot":     {table: "health_snapshots", hasProject: true},
	"contract":            {table: "contracts", hasProject: true},
	"field_inspection":    {table: "field_inspections", hasProject: true},
	"benefit_indicator":   {table: "benefit_indicators", hasProject: true},
	"benefit_measurement": {table: "benefit_measurements", hasProject: false},
}

func NewCommandHandler(db *gorm.DB, writer *audit.Writer) *CommandHandler {
	return &CommandHandler{db: db, audit: writer}
}
func commandClaims(c *gin.Context) (*auth.Claims, bool) {
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
func (h *CommandHandler) ListEscalations(c *gin.Context) {
	cl, ok := commandClaims(c)
	if !ok {
		return
	}
	var v []Escalation
	err := h.db.Where("organization_id=? AND deleted_at IS NULL", cl.OrganizationID).Order("created_at DESC").Find(&v).Error
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "escalations retrieved", v)
}
func (h *CommandHandler) CreateEscalation(c *gin.Context) {
	cl, ok := commandClaims(c)
	if !ok {
		return
	}
	var req CreateEscalationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	projectID, err := h.validateCommandSource(cl.OrganizationID, req.ProjectID, req.SourceType, req.SourceID)
	if err != nil {
		h.respondValidationError(c, err)
		return
	}
	item := &Escalation{ID: uuid.New(), OrganizationID: cl.OrganizationID, ProjectID: projectID, SourceType: req.SourceType, SourceID: req.SourceID, Level: req.Level, Reason: req.Reason, Status: "OPEN", CreatedBy: cl.UserID, CreatedAt: time.Now()}
	if err := h.db.Create(item).Error; err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	h.record(cl, item.ID, "command_escalation.created")
	response.Created(c, "escalation created", item)
}
func (h *CommandHandler) ListDecisions(c *gin.Context) {
	cl, ok := commandClaims(c)
	if !ok {
		return
	}
	var v []Decision
	err := h.db.Where("organization_id=? AND deleted_at IS NULL", cl.OrganizationID).Order("due_date ASC NULLS LAST, created_at DESC").Find(&v).Error
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "executive decisions retrieved", v)
}
func (h *CommandHandler) CreateDecision(c *gin.Context) {
	cl, ok := commandClaims(c)
	if !ok {
		return
	}
	var req CreateDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.ProjectID != nil {
		if err := h.validateProject(cl.OrganizationID, *req.ProjectID); err != nil {
			h.respondValidationError(c, err)
			return
		}
	}
	if req.EscalationID != nil {
		escalationProjectID, err := h.validateEscalation(cl.OrganizationID, *req.EscalationID)
		if err != nil {
			h.respondValidationError(c, err)
			return
		}
		if req.ProjectID == nil {
			req.ProjectID = escalationProjectID
		} else if escalationProjectID != nil && *req.ProjectID != *escalationProjectID {
			response.NotFound(c, "escalation not found")
			return
		}
	}
	if req.OwnerID != nil {
		if err := h.validateUser(cl.OrganizationID, *req.OwnerID); err != nil {
			h.respondValidationError(c, err)
			return
		}
	}
	now := time.Now()
	item := &Decision{ID: uuid.New(), OrganizationID: cl.OrganizationID, ProjectID: req.ProjectID, EscalationID: req.EscalationID, Subject: req.Subject, DecisionText: req.DecisionText, OwnerID: req.OwnerID, DueDate: req.DueDate, Status: "OPEN", DecidedBy: &cl.UserID, DecidedAt: &now, CreatedBy: cl.UserID, CreatedAt: now}
	if err := h.db.Create(item).Error; err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	h.record(cl, item.ID, "executive_decision.created")
	response.Created(c, "executive decision created", item)
}
func (h *CommandHandler) UpdateEscalationStatus(c *gin.Context) {
	cl, ok := commandClaims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("escalationID"))
	if err != nil {
		response.BadRequest(c, "invalid escalation id")
		return
	}
	var req StatusRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.Status != "ACKNOWLEDGED" && req.Status != "CLOSED" {
		response.BadRequest(c, "status must be ACKNOWLEDGED or CLOSED")
		return
	}
	var item Escalation
	if err = h.db.Where("organization_id=? AND id=? AND deleted_at IS NULL", cl.OrganizationID, id).First(&item).Error; err != nil {
		response.NotFound(c, "escalation not found")
		return
	}
	item.Status = req.Status
	if req.Status == "ACKNOWLEDGED" {
		now := time.Now()
		item.AcknowledgedBy = &cl.UserID
		item.AcknowledgedAt = &now
	}
	if err = h.db.Save(&item).Error; err != nil {
		response.InternalError(c)
		return
	}
	h.record(cl, id, "command_escalation."+strings.ToLower(req.Status))
	response.OK(c, "escalation status updated", item)
}
func (h *CommandHandler) UpdateDecisionStatus(c *gin.Context) {
	cl, ok := commandClaims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("decisionID"))
	if err != nil {
		response.BadRequest(c, "invalid decision id")
		return
	}
	var req StatusRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.Status != "IN_PROGRESS" && req.Status != "COMPLETED" && req.Status != "CANCELLED" {
		response.BadRequest(c, "unsupported decision status")
		return
	}
	var item Decision
	if err = h.db.Where("organization_id=? AND id=? AND deleted_at IS NULL", cl.OrganizationID, id).First(&item).Error; err != nil {
		response.NotFound(c, "decision not found")
		return
	}
	item.Status = req.Status
	if req.Status == "COMPLETED" || req.Status == "CANCELLED" {
		now := time.Now()
		item.DecidedBy = &cl.UserID
		item.DecidedAt = &now
	}
	if err = h.db.Save(&item).Error; err != nil {
		response.InternalError(c)
		return
	}
	h.record(cl, id, "executive_decision."+strings.ToLower(req.Status))
	response.OK(c, "decision status updated", item)
}
func (h *CommandHandler) record(cl *auth.Claims, id uuid.UUID, action string) {
	if h.audit != nil {
		h.audit.Record(audit.WriteRequest{OrganizationID: cl.OrganizationID, ActorID: &cl.UserID, Action: action, EntityType: "command_center", EntityID: id.String()})
	}
}

var errCommandNotFound = errors.New("command resource not found")

func (h *CommandHandler) validateProject(orgID, projectID uuid.UUID) error {
	var count int64
	if err := h.db.Table("projects").Where("id = ? AND organization_id = ? AND deleted_at IS NULL", projectID, orgID).Count(&count).Error; err != nil {
		return err
	}
	if count != 1 {
		return errCommandNotFound
	}
	return nil
}

func (h *CommandHandler) validateUser(orgID, userID uuid.UUID) error {
	var count int64
	if err := h.db.Table("users").Where("id = ? AND organization_id = ? AND deleted_at IS NULL", userID, orgID).Count(&count).Error; err != nil {
		return err
	}
	if count != 1 {
		return errCommandNotFound
	}
	return nil
}

func (h *CommandHandler) validateEscalation(orgID, escalationID uuid.UUID) (*uuid.UUID, error) {
	var row struct {
		ProjectID *uuid.UUID
	}
	if err := h.db.Table("command_escalations").Select("project_id").Where("id = ? AND organization_id = ? AND deleted_at IS NULL", escalationID, orgID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errCommandNotFound
		}
		return nil, err
	}
	return row.ProjectID, nil
}

func (h *CommandHandler) validateCommandSource(orgID uuid.UUID, requestedProjectID *uuid.UUID, sourceType string, sourceID uuid.UUID) (*uuid.UUID, error) {
	source, ok := commandSources[sourceType]
	if !ok {
		return nil, errors.New("unsupported source_type")
	}
	if sourceType == "project" {
		if err := h.validateProject(orgID, sourceID); err != nil {
			return nil, err
		}
		if requestedProjectID != nil && *requestedProjectID != sourceID {
			return nil, errCommandNotFound
		}
		projectID := sourceID
		return &projectID, nil
	}

	if source.hasProject {
		var row struct {
			ProjectID uuid.UUID
		}
		if err := h.db.Table(source.table).Select("project_id").Where("id = ? AND organization_id = ? AND deleted_at IS NULL", sourceID, orgID).First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errCommandNotFound
			}
			return nil, err
		}
		if requestedProjectID != nil && *requestedProjectID != row.ProjectID {
			return nil, errCommandNotFound
		}
		return &row.ProjectID, nil
	}

	var count int64
	if err := h.db.Table(source.table).Where("id = ? AND organization_id = ? AND deleted_at IS NULL", sourceID, orgID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count != 1 {
		return nil, errCommandNotFound
	}
	if requestedProjectID != nil {
		if err := h.validateProject(orgID, *requestedProjectID); err != nil {
			return nil, err
		}
	}
	return requestedProjectID, nil
}

func (h *CommandHandler) respondValidationError(c *gin.Context, err error) {
	if errors.Is(err, errCommandNotFound) {
		response.NotFound(c, "command source not found")
		return
	}
	response.BadRequest(c, err.Error())
}
