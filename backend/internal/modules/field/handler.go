package field

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"go.uber.org/zap"
)

type Handler struct {
	svc *Service
	log *zap.Logger
}

func NewHandler(svc *Service, log *zap.Logger) *Handler { return &Handler{svc: svc, log: log} }
func (h *Handler) claims(c *gin.Context) (*auth.Claims, bool) {
	value, ok := c.Get(string(auth.ContextKeyClaims))
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return nil, false
	}
	claims, ok := value.(*auth.Claims)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return nil, false
	}
	return claims, true
}
func projectID(c *gin.Context) (uuid.UUID, error) { return uuid.Parse(c.Param("id")) }

func (h *Handler) List(c *gin.Context) {
	claims, ok := h.claims(c)
	if !ok {
		return
	}
	pid, err := projectID(c)
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}
	items, err := h.svc.List(claims.OrganizationID, pid)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "project not found")
		return
	}
	if err != nil {
		h.log.Error("field: list inspections", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, "field inspections retrieved", items)
}
func (h *Handler) Create(c *gin.Context) {
	claims, ok := h.claims(c)
	if !ok {
		return
	}
	pid, err := projectID(c)
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}
	var req CreateInspectionRequest
	if err = c.ShouldBind(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	file, _ := c.FormFile("file")
	item, err := h.svc.Create(claims.OrganizationID, claims.UserID, pid, req, file)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "project not found")
		return
	}
	if errors.Is(err, ErrInvalidEvidence) {
		response.BadRequest(c, "evidence file type is not allowed")
		return
	}
	if err != nil {
		h.log.Error("field: create inspection", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "field inspection created", item)
}
func (h *Handler) Verify(c *gin.Context) {
	claims, ok := h.claims(c)
	if !ok {
		return
	}
	pid, err := projectID(c)
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}
	iid, err := uuid.Parse(c.Param("inspectionID"))
	if err != nil {
		response.BadRequest(c, "invalid inspection id")
		return
	}
	var req VerifyRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.svc.Verify(claims.OrganizationID, claims.UserID, pid, iid, req)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "inspection not found")
		return
	}
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "inspection verification updated", item)
}
func (h *Handler) Delete(c *gin.Context) {
	claims, ok := h.claims(c)
	if !ok {
		return
	}
	pid, err := projectID(c)
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}
	iid, err := uuid.Parse(c.Param("inspectionID"))
	if err != nil {
		response.BadRequest(c, "invalid inspection id")
		return
	}
	if err = h.svc.Delete(claims.OrganizationID, claims.UserID, pid, iid); err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(c, "inspection not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.NoContent(c)
}
func (h *Handler) Download(c *gin.Context) {
	claims, ok := h.claims(c)
	if !ok {
		return
	}
	pid, err := projectID(c)
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}
	iid, err := uuid.Parse(c.Param("inspectionID"))
	if err != nil {
		response.BadRequest(c, "invalid inspection id")
		return
	}
	eid, err := uuid.Parse(c.Param("evidenceID"))
	if err != nil {
		response.BadRequest(c, "invalid evidence id")
		return
	}
	item, path, err := h.svc.OpenEvidence(claims.OrganizationID, pid, iid, eid)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "evidence not found")
		return
	}
	if err != nil {
		h.log.Error("field: download evidence", zap.Error(err))
		response.InternalError(c)
		return
	}
	name := strings.ReplaceAll(item.FileName, "\"", "'")
	c.Header("Content-Type", item.MimeType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", name, url.PathEscape(item.FileName)))
	c.Header("X-Content-Type-Options", "nosniff")
	c.File(path)
}

func (h *Handler) AddEvidence(c *gin.Context) {
	claims, ok := h.claims(c)
	if !ok {
		return
	}
	pid, err := projectID(c)
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}
	iid, err := uuid.Parse(c.Param("inspectionID"))
	if err != nil {
		response.BadRequest(c, "invalid inspection id")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}
	// Optional per-evidence geotag from form fields
	var lat, lon *float64
	if v := c.PostForm("latitude"); v != "" {
		var f float64
		if _, scanErr := fmt.Sscanf(v, "%f", &f); scanErr == nil {
			lat = &f
		}
	}
	if v := c.PostForm("longitude"); v != "" {
		var f float64
		if _, scanErr := fmt.Sscanf(v, "%f", &f); scanErr == nil {
			lon = &f
		}
	}
	item, err := h.svc.AddEvidence(claims.OrganizationID, claims.UserID, pid, iid, file, lat, lon)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "inspection not found")
		return
	}
	if errors.Is(err, ErrInvalidEvidence) {
		response.BadRequest(c, "evidence file type is not allowed")
		return
	}
	if err != nil {
		h.log.Error("field: add evidence", zap.Error(err))
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "evidence added", item)
}
