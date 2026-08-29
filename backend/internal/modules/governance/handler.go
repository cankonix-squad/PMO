package governance

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"go.uber.org/zap"
)

// bindOptionalJSON binds a JSON body that may be empty. An empty body (EOF) is
// allowed; malformed JSON returns an error so the handler can answer 400.
func bindOptionalJSON(c *gin.Context, dst interface{}) error {
	if c.Request.Body == nil {
		return nil
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return err
	}
	return nil
}

// Handler exposes data governance endpoints.
type Handler struct {
	svc *Service
	log *zap.Logger
}

// NewHandler creates a governance Handler.
func NewHandler(svc *Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// claims extracts JWT claims from gin context.
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

// RegisterRoutes wires all governance routes onto the provided RouterGroup.
// The group should already have AuthRequired middleware applied.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	// Submissions
	rg.GET("/submissions", h.ListSubmissions)
	rg.POST("/submissions", h.CreateSubmission)
	rg.GET("/submissions/:id", h.GetSubmission)
	rg.POST("/submissions/:id/submit", h.Submit)
	rg.POST("/submissions/:id/review", h.StartReview)
	rg.POST("/submissions/:id/approve", h.Approve)
	rg.POST("/submissions/:id/reject", h.Reject)
	rg.POST("/submissions/:id/lock", h.Lock)
	rg.POST("/submissions/:id/cancel", h.Cancel)

	// Lock periods
	rg.GET("/lock-periods", h.ListLockPeriods)
	rg.POST("/lock-periods", h.CreateLockPeriod)
	rg.POST("/lock-periods/:id/lock", h.LockPeriod)
}

// ---------------------------------------------------------------------------
// GET /api/v1/governance/submissions
// ---------------------------------------------------------------------------

func (h *Handler) ListSubmissions(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	f := ListSubmissionsFilter{
		Status:      c.Query("status"),
		DatasetType: c.Query("dataset_type"),
		SourceType:  c.Query("source_type"),
	}
	f.PeriodYear, _ = strconv.Atoi(c.Query("period_year"))
	f.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	f.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.svc.ListSubmissions(cl.OrganizationID, f)
	if err != nil {
		h.log.Error("governance: list submissions", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, "submissions retrieved", gin.H{
		"data": list,
		"meta": gin.H{"total": total, "page": f.Page, "page_size": f.PageSize},
	})
}

// ---------------------------------------------------------------------------
// POST /api/v1/governance/submissions
// ---------------------------------------------------------------------------

func (h *Handler) CreateSubmission(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	var req CreateSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	sub, err := h.svc.CreateSubmission(cl.OrganizationID, cl.UserID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidDatasetType):
			response.BadRequest(c, "invalid dataset_type")
		case errors.Is(err, ErrInvalidSourceType):
			response.BadRequest(c, "invalid source_type")
		case errors.Is(err, ErrInvalidEntityType):
			response.BadRequest(c, err.Error())
		case errors.Is(err, ErrEntityNotFound):
			response.BadRequest(c, err.Error())
		case errors.Is(err, ErrEntityDeleted):
			response.BadRequest(c, err.Error())
		case errors.Is(err, ErrPendingMatchNotReady):
			response.BadRequest(c, err.Error())
		case errors.Is(err, ErrMappingRejected):
			response.BadRequest(c, err.Error())
		case errors.Is(err, ErrEmptyItems):
			response.BadRequest(c, "submission must contain at least one item")
		case errors.Is(err, ErrLockedPeriod):
			response.Conflict(c, "period is locked; cannot create submission")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.Created(c, "submission created", sub)
}

// ---------------------------------------------------------------------------
// GET /api/v1/governance/submissions/:id
// ---------------------------------------------------------------------------

func (h *Handler) GetSubmission(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid submission id")
		return
	}
	detail, err := h.svc.GetSubmission(cl.OrganizationID, id)
	if errors.Is(err, ErrSubmissionNotFound) {
		response.NotFound(c, "submission not found")
		return
	}
	if err != nil {
		h.log.Error("governance: get submission", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, "submission retrieved", detail)
}

// ---------------------------------------------------------------------------
// POST /api/v1/governance/submissions/:id/submit
// ---------------------------------------------------------------------------

func (h *Handler) Submit(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid submission id")
		return
	}
	sub, err := h.svc.Submit(cl.OrganizationID, cl.UserID, id)
	if err != nil {
		h.transitionError(c, err)
		return
	}
	response.OK(c, "submission submitted", sub)
}

// ---------------------------------------------------------------------------
// POST /api/v1/governance/submissions/:id/review
// ---------------------------------------------------------------------------

func (h *Handler) StartReview(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid submission id")
		return
	}
	var req ReviewRequest
	if err := bindOptionalJSON(c, &req); err != nil {
		response.BadRequest(c, "invalid JSON: "+err.Error())
		return
	}
	detail, err := h.svc.StartReview(cl.OrganizationID, cl.UserID, id, req)
	if err != nil {
		h.transitionError(c, err)
		return
	}
	response.OK(c, "review started", detail)
}

// ---------------------------------------------------------------------------
// POST /api/v1/governance/submissions/:id/approve
// ---------------------------------------------------------------------------

func (h *Handler) Approve(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid submission id")
		return
	}
	sub, err := h.svc.Approve(cl.OrganizationID, cl.UserID, id)
	if err != nil {
		h.transitionError(c, err)
		return
	}
	response.OK(c, "submission approved", sub)
}

// ---------------------------------------------------------------------------
// POST /api/v1/governance/submissions/:id/reject
// ---------------------------------------------------------------------------

func (h *Handler) Reject(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid submission id")
		return
	}
	var req RejectRequest
	if err := bindOptionalJSON(c, &req); err != nil {
		response.BadRequest(c, "invalid JSON: "+err.Error())
		return
	}
	sub, err := h.svc.Reject(cl.OrganizationID, cl.UserID, id, req)
	if err != nil {
		h.transitionError(c, err)
		return
	}
	response.OK(c, "submission rejected", sub)
}

// ---------------------------------------------------------------------------
// POST /api/v1/governance/submissions/:id/lock
// ---------------------------------------------------------------------------

func (h *Handler) Lock(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid submission id")
		return
	}
	var req LockRequest
	if err := bindOptionalJSON(c, &req); err != nil {
		response.BadRequest(c, "invalid JSON: "+err.Error())
		return
	}
	sub, err := h.svc.Lock(cl.OrganizationID, cl.UserID, id, req)
	if err != nil {
		h.transitionError(c, err)
		return
	}
	response.OK(c, "submission locked", sub)
}

// ---------------------------------------------------------------------------
// POST /api/v1/governance/submissions/:id/cancel
// ---------------------------------------------------------------------------

func (h *Handler) Cancel(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid submission id")
		return
	}
	var req CancelRequest
	if err := bindOptionalJSON(c, &req); err != nil {
		response.BadRequest(c, "invalid JSON: "+err.Error())
		return
	}
	sub, err := h.svc.Cancel(cl.OrganizationID, cl.UserID, id, req)
	if err != nil {
		h.transitionError(c, err)
		return
	}
	response.OK(c, "submission cancelled", sub)
}

// ---------------------------------------------------------------------------
// GET /api/v1/governance/lock-periods
// ---------------------------------------------------------------------------

func (h *Handler) ListLockPeriods(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	f := ListLockPeriodsFilter{
		DatasetType: c.Query("dataset_type"),
		Status:      c.Query("status"),
	}
	f.PeriodYear, _ = strconv.Atoi(c.Query("period_year"))
	f.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	f.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.svc.ListLockPeriods(cl.OrganizationID, f)
	if err != nil {
		h.log.Error("governance: list lock periods", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, "lock periods retrieved", gin.H{
		"data": list,
		"meta": gin.H{"total": total, "page": f.Page, "page_size": f.PageSize},
	})
}

// ---------------------------------------------------------------------------
// POST /api/v1/governance/lock-periods
// ---------------------------------------------------------------------------

func (h *Handler) CreateLockPeriod(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	var req CreateLockPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	period, err := h.svc.CreateLockPeriod(cl.OrganizationID, cl.UserID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidDatasetType):
			response.BadRequest(c, "invalid dataset_type")
		case errors.Is(err, ErrLockPeriodConflict):
			response.Conflict(c, "lock period already exists for this dataset/period")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.Created(c, "lock period created", period)
}

// ---------------------------------------------------------------------------
// POST /api/v1/governance/lock-periods/:id/lock
// ---------------------------------------------------------------------------

func (h *Handler) LockPeriod(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid lock period id")
		return
	}
	var req LockRequest
	if err := bindOptionalJSON(c, &req); err != nil {
		response.BadRequest(c, "invalid JSON: "+err.Error())
		return
	}
	period, err := h.svc.LockPeriodByID(cl.OrganizationID, cl.UserID, id, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrLockPeriodNotFound):
			response.NotFound(c, "lock period not found")
		case errors.Is(err, ErrInvalidTransition):
			response.Conflict(c, "lock period is already locked")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.OK(c, "lock period locked", period)
}

// transitionError maps service errors to appropriate HTTP status codes.
func (h *Handler) transitionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrSubmissionNotFound):
		response.NotFound(c, "submission not found")
	case errors.Is(err, ErrInvalidTransition):
		response.Conflict(c, "invalid submission status transition")
	case errors.Is(err, ErrRejectionReason):
		response.BadRequest(c, "rejection_reason is required")
	case errors.Is(err, ErrInvalidItems):
		response.BadRequest(c, "all items must be VALID before approving")
	case errors.Is(err, ErrEmptyItems):
		response.BadRequest(c, "submission has no items")
	case errors.Is(err, ErrLockedPeriod):
		response.Conflict(c, "period is locked; cannot modify")
	case errors.Is(err, ErrInvalidEntityType):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrEntityNotFound):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrEntityDeleted):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrPendingMatchNotReady):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrMappingRejected):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrUnknownMatchStatus):
		response.BadRequest(c, err.Error())
	default:
		response.BadRequest(c, err.Error())
	}
}
