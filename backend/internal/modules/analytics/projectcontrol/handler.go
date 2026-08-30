package projectcontrol

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"strconv"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }
func (h *Handler) Get(c *gin.Context) {
	v, ok := c.Get(string(auth.ContextKeyClaims))
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	cl, ok := v.(*auth.Claims)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	pid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}
	year, _ := strconv.Atoi(c.Query("year"))
	month, _ := strconv.Atoi(c.Query("month"))
	item, err := h.svc.Get(c.Request.Context(), cl.OrganizationID, pid, year, month)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "project control not found")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "project control retrieved", item)
}
