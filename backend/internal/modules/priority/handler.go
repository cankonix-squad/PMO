package priority

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
)

// Handler exposes priority scoring endpoints.
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

// GET /api/v1/priority/formulas
func (h *Handler) ListFormulas(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	v, err := h.svc.ListFormulas(cl.OrganizationID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "priority formulas retrieved", v)
}

// POST /api/v1/priority/formulas
func (h *Handler) CreateFormula(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	var r CreateFormulaRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	f, err := h.svc.CreateFormula(cl.OrganizationID, cl.UserID, r)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "priority formula created", f)
}

// GET /api/v1/priority/formulas/:id
func (h *Handler) GetFormula(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid formula id")
		return
	}
	f, err := h.svc.GetFormula(cl.OrganizationID, id)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "priority formula not found")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "priority formula retrieved", f)
}

// PUT /api/v1/priority/formulas/:id
func (h *Handler) UpdateFormula(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid formula id")
		return
	}
	var r UpdateFormulaRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	f, err := h.svc.UpdateFormula(cl.OrganizationID, cl.UserID, id, r)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "priority formula not found")
		return
	}
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "priority formula updated", f)
}

// POST /api/v1/priority/formulas/:id/activate
func (h *Handler) ActivateFormula(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid formula id")
		return
	}
	f, err := h.svc.ActivateFormula(cl.OrganizationID, cl.UserID, id)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "priority formula not found")
		return
	}
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "priority formula activated", f)
}

// POST /api/v1/priority/calculate
// Calculate for a single project: body must include project_id.
func (h *Handler) Calculate(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	var body struct {
		ProjectID string `json:"project_id" binding:"required"`
		FormulaID string `json:"formula_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	projectID, err := uuid.Parse(body.ProjectID)
	if err != nil {
		response.BadRequest(c, "invalid project_id")
		return
	}
	sc, err := h.svc.Calculate(cl.OrganizationID, cl.UserID, projectID, CalculateRequest{FormulaID: body.FormulaID})
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "project or formula not found")
		return
	}
	if errors.Is(err, ErrNoActiveFormula) {
		response.BadRequest(c, ErrNoActiveFormula.Error())
		return
	}
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "priority score calculated", sc)
}

// POST /api/v1/priority/batch-calculate
func (h *Handler) BatchCalculate(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	var r BatchCalculateRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		// body is optional — allow empty body
		r = BatchCalculateRequest{}
	}
	res, err := h.svc.BatchCalculate(cl.OrganizationID, cl.UserID, r)
	if errors.Is(err, ErrNoActiveFormula) {
		response.BadRequest(c, ErrNoActiveFormula.Error())
		return
	}
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "batch priority calculation complete", res)
}

// GET /api/v1/priority/projects
func (h *Handler) ListRanking(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	category := c.Query("category")
	formulaID := c.Query("formula_id")
	res, err := h.svc.ListRanking(cl.OrganizationID, category, formulaID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "priority ranking retrieved", res)
}

// GET /api/v1/priority/projects/:projectID
func (h *Handler) GetProjectScore(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	projectID, err := uuid.Parse(c.Param("projectID"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}
	sc, err := h.svc.GetScore(cl.OrganizationID, projectID)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "no priority score found for this project")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "project priority score retrieved", sc)
}

// GET /api/v1/priority/projects/:projectID/explain
// Returns the latest score with full component breakdown (same as GetProjectScore but explicit).
func (h *Handler) ExplainProjectScore(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	projectID, err := uuid.Parse(c.Param("projectID"))
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}
	sc, err := h.svc.GetScore(cl.OrganizationID, projectID)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "no priority score found for this project")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "priority score explanation retrieved", sc)
}
