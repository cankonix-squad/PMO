package health

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }
func claims(c *gin.Context) (*auth.Claims, bool) {
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
func (h *Handler) ListFormulas(c *gin.Context) {
	cl, ok := claims(c)
	if !ok {
		return
	}
	v, err := h.svc.ListFormulas(cl.OrganizationID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "health formulas retrieved", v)
}
func (h *Handler) CreateFormula(c *gin.Context) {
	cl, ok := claims(c)
	if !ok {
		return
	}
	var req CreateFormulaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	v, err := h.svc.CreateFormula(cl.OrganizationID, cl.UserID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "health formula created", v)
}
func (h *Handler) TransitionFormula(c *gin.Context) {
	cl, ok := claims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("formulaID"))
	if err != nil {
		response.BadRequest(c, "invalid formula id")
		return
	}
	var req FormulaTransitionRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	v, err := h.svc.TransitionFormula(cl.OrganizationID, cl.UserID, id, req.Status)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "formula not found")
		return
	}
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "health formula status updated", v)
}
func (h *Handler) Calculate(c *gin.Context) {
	cl, ok := claims(c)
	if !ok {
		return
	}
	pid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}
	var req CalculateRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	v, err := h.svc.Calculate(cl.OrganizationID, cl.UserID, pid, req)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "project or formula not found")
		return
	}
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "health snapshot calculated", v)
}
func (h *Handler) ListSnapshots(c *gin.Context) {
	cl, ok := claims(c)
	if !ok {
		return
	}
	pid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}
	v, err := h.svc.ListSnapshots(cl.OrganizationID, pid)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "health snapshots retrieved", v)
}
