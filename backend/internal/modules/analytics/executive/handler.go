package executive

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func claimsFromCtx(c *gin.Context) (*auth.Claims, bool) {
	v, ok := c.Get(string(auth.ContextKeyClaims))
	if !ok {
		return nil, false
	}
	cl, ok := v.(*auth.Claims)
	return cl, ok
}

func filterFromQuery(c *gin.Context) Filter {
	year, _ := strconv.Atoi(c.Query("year"))
	month, _ := strconv.Atoi(c.Query("month"))
	return Filter{PeriodYear: year, PeriodMonth: month}
}

// GET /analytics/executive
func (h *Handler) GetDashboard(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	res, err := h.svc.GetDashboard(c.Request.Context(), cl.OrganizationID, filterFromQuery(c))
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "executive dashboard retrieved", res)
}
