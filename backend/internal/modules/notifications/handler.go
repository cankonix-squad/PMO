package notifications

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/core/notification"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"github.com/harmanto-49/cankora/internal/shared/constants"
	"github.com/harmanto-49/cankora/internal/shared/types"
)

// Handler exposes notification endpoints.
type Handler struct {
	svc *notification.Service
}

// NewHandler creates a Handler.
func NewHandler(svc *notification.Service) *Handler {
	return &Handler{svc: svc}
}

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

func (h *Handler) buildListFilter(c *gin.Context, orgID uuid.UUID, recipientID *uuid.UUID) notification.ListFilter {
	filter := notification.ListFilter{
		OrganizationID:  orgID,
		RecipientUserID: recipientID,
		Status:          c.Query("status"),
		Channel:         c.Query("channel"),
		Priority:        c.Query("priority"),
		SourceType:      c.Query("source_type"),
		UnreadOnly:      c.Query("unread_only") == "true",
	}
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = constants.DefaultPageSize
	}
	if pageSize > constants.MaxPageSize {
		pageSize = constants.MaxPageSize
	}
	filter.Page = page
	filter.PageSize = pageSize
	return filter
}

// GET /api/v1/notifications
// Users see their own notifications.
func (h *Handler) List(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	filter := h.buildListFilter(c, cl.OrganizationID, &cl.UserID)
	items, total, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c)
		return
	}
	meta := types.NewPaginationMeta(filter.Page, filter.PageSize, total)
	response.OKPaginated(c, "notifications retrieved", items, meta)
}

// GET /api/v1/notifications/admin — all org notifications (admin/PMO only)
func (h *Handler) ListAdmin(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	// No recipient filter — admin sees all in org
	filter := h.buildListFilter(c, cl.OrganizationID, nil)
	items, total, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c)
		return
	}
	meta := types.NewPaginationMeta(filter.Page, filter.PageSize, total)
	response.OKPaginated(c, "notifications retrieved", items, meta)
}

// GET /api/v1/notifications/summary
func (h *Handler) Summary(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	s, err := h.svc.Summary(c.Request.Context(), cl.OrganizationID, &cl.UserID)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "notification summary retrieved", s)
}

// GET /api/v1/notifications/:id
func (h *Handler) GetByID(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid notification id")
		return
	}
	n, err := h.svc.GetByID(c.Request.Context(), cl.OrganizationID, id)
	if err != nil {
		if err == notification.ErrNotFound {
			response.NotFound(c, "notification not found")
			return
		}
		response.InternalError(c)
		return
	}
	// Users can only read their own notifications.
	if n.RecipientUserID == nil || *n.RecipientUserID != cl.UserID {
		response.NotFound(c, "notification not found")
		return
	}
	response.OK(c, "notification retrieved", n)
}

// PATCH /api/v1/notifications/:id/read
func (h *Handler) MarkRead(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid notification id")
		return
	}
	// Verify ownership before marking read
	n, err := h.svc.GetByID(c.Request.Context(), cl.OrganizationID, id)
	if err != nil || n.RecipientUserID == nil || *n.RecipientUserID != cl.UserID {
		response.NotFound(c, "notification not found")
		return
	}
	if err := h.svc.MarkRead(c.Request.Context(), cl.OrganizationID, id); err != nil {
		if err == notification.ErrNotFound {
			response.NotFound(c, "notification not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, "notification marked as read", nil)
}

// PATCH /api/v1/notifications/read-all
func (h *Handler) MarkAllRead(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	if err := h.svc.MarkAllRead(c.Request.Context(), cl.OrganizationID, cl.UserID); err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "all notifications marked as read", nil)
}

// POST /api/v1/notifications/:id/retry
func (h *Handler) Retry(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid notification id")
		return
	}
	if err := h.svc.Retry(c.Request.Context(), cl.OrganizationID, id); err != nil {
		if err == notification.ErrNotFound {
			response.NotFound(c, "notification not found")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "notification retry enqueued", nil)
}

// TestNotificationRequest is the body for POST /notifications/test.
type TestNotificationRequest struct {
	Subject    string `json:"subject"     binding:"required"`
	Body       string `json:"body"        binding:"required"`
	Channel    string `json:"channel"`
	Priority   string `json:"priority"`
	SourceType string `json:"source_type"`
}

// POST /api/v1/notifications/test
func (h *Handler) CreateTest(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}
	var req TestNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	channel := req.Channel
	if channel == "" {
		channel = notification.ChannelInApp
	}
	priority := req.Priority
	if priority == "" {
		priority = notification.PriorityNormal
	}
	enqReq := notification.EnqueueRequest{
		OrganizationID:  cl.OrganizationID,
		RecipientUserID: &cl.UserID,
		Channel:         channel,
		Priority:        priority,
		Subject:         req.Subject,
		Body:            req.Body,
		SourceType:      req.SourceType,
		SourceID:        "test",
	}
	n, err := h.svc.EnqueueAndReturn(enqReq)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Created(c, "test notification created", n)
}

// RegisterRoutes attaches all user-facing notification routes.
// Caller must apply auth + permission middleware.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.List)
	rg.GET("/summary", h.Summary)
	rg.POST("/test", h.CreateTest)
	rg.PATCH("/read-all", h.MarkAllRead)
	rg.GET("/admin", h.ListAdmin)
	rg.GET("/:id", h.GetByID)
	rg.PATCH("/:id/read", h.MarkRead)
	rg.POST("/:id/retry", h.Retry)
}
