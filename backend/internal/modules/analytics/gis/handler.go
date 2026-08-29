package gis

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
)

// Handler melayani endpoint GIS Map (P2-008).
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

func filterFromQuery(c *gin.Context) GISFilter {
	return GISFilter{
		Status:      c.Query("status"),
		HealthClass: c.Query("health_class"),
		Province:    c.Query("province"),
	}
}

// GET /analytics/gis/projects
// Mengembalikan semua marker proyek (dengan atau tanpa koordinat) untuk peta.
func (h *Handler) GetProjects(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	markers, err := h.svc.GetProjects(c.Request.Context(), cl.OrganizationID, filterFromQuery(c))
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "gis projects retrieved", markers)
}

// GET /analytics/gis/summary
// Mengembalikan ringkasan statistik untuk panel samping peta.
func (h *Handler) GetSummary(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	summary, err := h.svc.GetSummary(c.Request.Context(), cl.OrganizationID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "gis summary retrieved", summary)
}

// GET /analytics/gis/projects/:id
// Mengembalikan detail satu proyek untuk popup marker.
func (h *Handler) GetProjectDetail(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}
	detail, err := h.svc.GetProjectDetail(c.Request.Context(), cl.OrganizationID, projectID)
	if err != nil {
		response.InternalError(c)
		return
	}
	if detail == nil {
		response.NotFound(c, "project not found")
		return
	}
	response.OK(c, "gis project detail retrieved", detail)
}
