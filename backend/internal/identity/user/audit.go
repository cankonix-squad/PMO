package user

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/harmanto-49/cankora/internal/governance/audit"
	"github.com/harmanto-49/cankora/internal/identity/auth"
)

// WithAudit attaches an audit writer so lifecycle events are recorded.
func (h *Handler) WithAudit(w *audit.Writer) *Handler {
	h.audit = w
	return h
}

// recordAudit writes an asynchronous audit entry. orgID/actor come from the
// given claims; a nil writer or nil claims makes this a no-op.
func (h *Handler) recordAudit(c *gin.Context, claims *auth.Claims, action, entityType string, entityID uuid.UUID, label string, newValues interface{}) {
	if h.audit == nil || claims == nil {
		return
	}
	actorID := claims.UserID
	h.audit.Record(audit.WriteRequest{
		OrganizationID: claims.OrganizationID,
		ActorID:        &actorID,
		ActorEmail:     claims.Email,
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
