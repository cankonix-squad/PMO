package dataquality

import (
	"errors"

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
	value, exists := c.Get(string(auth.ContextKeyClaims))
	if !exists {
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

func (h *Handler) List(c *gin.Context) {
	claims, ok := h.claims(c)
	if !ok {
		return
	}
	items, err := h.svc.List(claims.OrganizationID, c.Query("status"))
	if err != nil {
		h.log.Error("dataquality: list submissions", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, "validation queue retrieved", items)
}

func (h *Handler) Create(c *gin.Context) {
	claims, ok := h.claims(c)
	if !ok {
		return
	}
	var req CreateSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.svc.Create(claims.OrganizationID, claims.UserID, req)
	if err != nil {
		if errors.Is(err, ErrSubmissionNotFound) {
			response.NotFound(c, "snapshot not found")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "validation submission created", item)
}

func (h *Handler) Get(c *gin.Context) {
	claims, ok := h.claims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("submissionID"))
	if err != nil {
		response.BadRequest(c, "invalid submission id")
		return
	}
	item, err := h.svc.Get(claims.OrganizationID, id)
	if err != nil {
		if errors.Is(err, ErrSubmissionNotFound) {
			response.NotFound(c, "validation submission not found")
			return
		}
		h.log.Error("dataquality: get submission", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, "validation submission retrieved", item)
}

func (h *Handler) Transition(c *gin.Context) {
	claims, ok := h.claims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("submissionID"))
	if err != nil {
		response.BadRequest(c, "invalid submission id")
		return
	}
	var req TransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	item, err := h.svc.Transition(claims.OrganizationID, claims.UserID, id, req)
	if err != nil {
		if errors.Is(err, ErrSubmissionNotFound) {
			response.NotFound(c, "validation submission not found")
			return
		}
		if errors.Is(err, ErrSelfValidation) {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "validation status updated", item)
}
