package program

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"gorm.io/gorm"
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

// GET /analytics/programs
func (h *Handler) ListPrograms(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	res, err := h.svc.ListPrograms(c.Request.Context(), cl.OrganizationID, filterFromQuery(c))
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "programs retrieved", res)
}

// GET /analytics/programs/:id
func (h *Handler) GetProgram(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid program id")
		return
	}
	res, err := h.svc.GetProgram(c.Request.Context(), cl.OrganizationID, id, filterFromQuery(c))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.NotFound(c, "program not found")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "program dashboard retrieved", res)
}

// GET /analytics/sectors
func (h *Handler) ListSectors(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	res, err := h.svc.ListSectors(c.Request.Context(), cl.OrganizationID, filterFromQuery(c))
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "sectors retrieved", res)
}

// GET /analytics/sectors/:id
func (h *Handler) GetSector(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid sector id")
		return
	}
	res, err := h.svc.GetSector(c.Request.Context(), cl.OrganizationID, id, filterFromQuery(c))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.NotFound(c, "sector not found")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "sector dashboard retrieved", res)
}
