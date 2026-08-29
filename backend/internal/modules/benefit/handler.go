package benefit

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
)

type Handler struct{ svc *Service }

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }
func (h *Handler) claims(c *gin.Context) (*auth.Claims, bool) {
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
func (h *Handler) List(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	v, e := h.svc.ListIndicators(cl.OrganizationID)
	if e != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "benefit indicators retrieved", v)
}
func (h *Handler) Get(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, e := uuid.Parse(c.Param("indicatorID"))
	if e != nil {
		response.BadRequest(c, "invalid indicator id")
		return
	}
	v, e := h.svc.GetIndicator(cl.OrganizationID, id)
	if errors.Is(e, ErrNotFound) {
		response.NotFound(c, "benefit indicator not found")
		return
	}
	if e != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "benefit indicator retrieved", v)
}
func (h *Handler) Create(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	var r CreateIndicatorRequest
	if e := c.ShouldBindJSON(&r); e != nil {
		response.BadRequest(c, e.Error())
		return
	}
	v, e := h.svc.CreateIndicator(cl.OrganizationID, cl.UserID, r)
	if errors.Is(e, ErrNotFound) {
		response.NotFound(c, "project not found")
		return
	}
	if e != nil {
		response.BadRequest(c, e.Error())
		return
	}
	response.Created(c, "benefit indicator created", v)
}
func (h *Handler) Update(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, e := uuid.Parse(c.Param("indicatorID"))
	if e != nil {
		response.BadRequest(c, "invalid indicator id")
		return
	}
	var r UpdateIndicatorRequest
	if e := c.ShouldBindJSON(&r); e != nil {
		response.BadRequest(c, e.Error())
		return
	}
	v, e := h.svc.UpdateIndicator(cl.OrganizationID, cl.UserID, id, r)
	if errors.Is(e, ErrNotFound) {
		response.NotFound(c, "benefit indicator not found")
		return
	}
	if e != nil {
		response.BadRequest(c, e.Error())
		return
	}
	response.OK(c, "benefit indicator updated", v)
}
func (h *Handler) Delete(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, e := uuid.Parse(c.Param("indicatorID"))
	if e != nil {
		response.BadRequest(c, "invalid indicator id")
		return
	}
	e = h.svc.DeleteIndicator(cl.OrganizationID, cl.UserID, id)
	if errors.Is(e, ErrNotFound) {
		response.NotFound(c, "benefit indicator not found")
		return
	}
	if e != nil {
		response.InternalError(c)
		return
	}
	response.NoContent(c)
}
func (h *Handler) ListMeasurements(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, e := uuid.Parse(c.Param("indicatorID"))
	if e != nil {
		response.BadRequest(c, "invalid indicator id")
		return
	}
	v, e := h.svc.ListMeasurements(cl.OrganizationID, id)
	if errors.Is(e, ErrNotFound) {
		response.NotFound(c, "benefit indicator not found")
		return
	}
	if e != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "benefit measurements retrieved", v)
}
func (h *Handler) CreateMeasurement(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, e := uuid.Parse(c.Param("indicatorID"))
	if e != nil {
		response.BadRequest(c, "invalid indicator id")
		return
	}
	var r CreateMeasurementRequest
	if e = c.ShouldBindJSON(&r); e != nil {
		response.BadRequest(c, e.Error())
		return
	}
	v, e := h.svc.CreateMeasurement(cl.OrganizationID, cl.UserID, id, r)
	if errors.Is(e, ErrNotFound) {
		response.NotFound(c, "indicator not found")
		return
	}
	if e != nil {
		response.BadRequest(c, e.Error())
		return
	}
	response.Created(c, "benefit measurement created", v)
}
func (h *Handler) UpdateMeasurement(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	indicatorID, e := uuid.Parse(c.Param("indicatorID"))
	if e != nil {
		response.BadRequest(c, "invalid indicator id")
		return
	}
	measurementID, e := uuid.Parse(c.Param("measurementID"))
	if e != nil {
		response.BadRequest(c, "invalid measurement id")
		return
	}
	var r UpdateMeasurementRequest
	if e = c.ShouldBindJSON(&r); e != nil {
		response.BadRequest(c, e.Error())
		return
	}
	v, e := h.svc.UpdateMeasurement(cl.OrganizationID, cl.UserID, indicatorID, measurementID, r)
	if errors.Is(e, ErrNotFound) {
		response.NotFound(c, "benefit measurement not found")
		return
	}
	if e != nil {
		response.BadRequest(c, e.Error())
		return
	}
	response.OK(c, "benefit measurement updated", v)
}
func (h *Handler) DeleteMeasurement(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	indicatorID, e := uuid.Parse(c.Param("indicatorID"))
	if e != nil {
		response.BadRequest(c, "invalid indicator id")
		return
	}
	measurementID, e := uuid.Parse(c.Param("measurementID"))
	if e != nil {
		response.BadRequest(c, "invalid measurement id")
		return
	}
	e = h.svc.DeleteMeasurement(cl.OrganizationID, cl.UserID, indicatorID, measurementID)
	if errors.Is(e, ErrNotFound) {
		response.NotFound(c, "benefit measurement not found")
		return
	}
	if e != nil {
		response.InternalError(c)
		return
	}
	response.NoContent(c)
}
func (h *Handler) Aggregate(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, e := uuid.Parse(c.Param("indicatorID"))
	if e != nil {
		response.BadRequest(c, "invalid indicator id")
		return
	}
	v, e := h.svc.Aggregate(cl.OrganizationID, id)
	if errors.Is(e, ErrNotFound) {
		response.NotFound(c, "indicator not found")
		return
	}
	if e != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "benefit aggregate retrieved", v)
}
func (h *Handler) Summary(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	v, e := h.svc.Summary(cl.OrganizationID)
	if e != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "benefit summary retrieved", v)
}
