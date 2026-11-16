package project

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/harmanto-49/cankora/internal/governance/audit"
)

// recordEntityAudit writes an asynchronous audit entry for a generic
// project-module entity (project, task, milestone).
func (h *Handler) recordEntityAudit(c *gin.Context, action, entityType string, entityID uuid.UUID, label string, newValues interface{}) {
	if h.audit == nil {
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
		EntityType:     entityType,
		EntityID:       entityID.String(),
		EntityLabel:    label,
		NewValues:      newValues,
		IPAddress:      c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		RequestID:      c.GetString("X-Request-ID"),
	})
}
