package commandcenter

import (
	"github.com/gin-gonic/gin"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }
func (h *Handler) Get(c *gin.Context) {
	v, ok := c.Get(string(auth.ContextKeyClaims))
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	claims, ok := v.(*auth.Claims)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	summary, err := h.svc.Summary(c.Request.Context(), claims.OrganizationID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "command center summary retrieved", summary)
}
