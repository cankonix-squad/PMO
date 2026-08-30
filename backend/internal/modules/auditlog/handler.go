package auditlog

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/audit"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"github.com/harmanto-49/cankora/internal/shared/constants"
	"github.com/harmanto-49/cankora/internal/shared/types"
)

// Handler exposes read-only audit log endpoints.
type Handler struct {
	repo audit.Repository
}

// NewHandler creates a Handler backed by the shared audit Repository.
func NewHandler(repo audit.Repository) *Handler {
	return &Handler{repo: repo}
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

// buildFilter reads query params and returns a ListFilter scoped to the caller's org.
func (h *Handler) buildFilter(c *gin.Context, orgID uuid.UUID) (audit.ListFilter, error) {
	f := audit.ListFilter{
		OrganizationID: orgID,
	}

	f.Action = c.Query("action")
	f.EntityType = c.Query("entity_type")
	f.EntityID = c.Query("entity_id")
	f.Search = c.Query("search")

	if raw := c.Query("actor_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return f, fmt.Errorf("invalid actor_id: %w", err)
		}
		f.ActorID = &id
	}

	if raw := c.Query("date_from"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			// try date-only
			t, err = time.Parse("2006-01-02", raw)
			if err != nil {
				return f, fmt.Errorf("invalid date_from (use RFC3339 or YYYY-MM-DD): %w", err)
			}
		}
		f.From = &t
	}

	if raw := c.Query("date_to"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			t, err = time.Parse("2006-01-02", raw)
			if err != nil {
				return f, fmt.Errorf("invalid date_to (use RFC3339 or YYYY-MM-DD): %w", err)
			}
		}
		// end-of-day inclusive when date-only
		t = t.Add(24*time.Hour - time.Second)
		f.To = &t
	}

	page := 1
	if raw := c.Query("page"); raw != "" {
		if p, err := strconv.Atoi(raw); err == nil && p > 0 {
			page = p
		}
	}
	pageSize := constants.DefaultPageSize
	if raw := c.Query("page_size"); raw != "" {
		if ps, err := strconv.Atoi(raw); err == nil && ps > 0 {
			if ps > constants.MaxPageSize {
				ps = constants.MaxPageSize
			}
			pageSize = ps
		}
	}
	f.Page = page
	f.PageSize = pageSize

	return f, nil
}

// GET /api/v1/audit-logs
func (h *Handler) List(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	filter, err := h.buildFilter(c, cl.OrganizationID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	logs, total, err := h.repo.List(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c)
		return
	}

	meta := types.NewPaginationMeta(filter.Page, filter.PageSize, total)
	response.OKPaginated(c, "audit logs retrieved", logs, meta)
}

// GET /api/v1/audit-logs/:id
func (h *Handler) GetByID(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid audit log id")
		return
	}

	entry, err := h.repo.GetByID(c.Request.Context(), cl.OrganizationID, id)
	if err != nil {
		if err == audit.ErrNotFound {
			response.NotFound(c, "audit log not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "audit log retrieved", entry)
}

// GET /api/v1/audit-logs/export  (CSV export — lightweight, max 1000 rows)
func (h *Handler) Export(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	filter, err := h.buildFilter(c, cl.OrganizationID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Cap export at 1000 rows regardless of pagination params
	filter.Page = 1
	filter.PageSize = 1000

	logs, _, err := h.repo.List(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c)
		return
	}

	filename := fmt.Sprintf("audit-logs-%s.csv", time.Now().Format("20060102-150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	w := csv.NewWriter(c.Writer)
	// Header row
	_ = w.Write([]string{
		"id", "created_at", "actor_email", "actor_id",
		"action", "entity_type", "entity_id", "entity_label",
		"ip_address", "request_id",
	})
	for _, l := range logs {
		actorID := ""
		if l.ActorID != nil {
			actorID = l.ActorID.String()
		}
		_ = w.Write([]string{
			l.ID.String(),
			l.CreatedAt.Format(time.RFC3339),
			l.ActorEmail,
			actorID,
			l.Action,
			l.EntityType,
			l.EntityID,
			l.EntityLabel,
			l.IPAddress,
			l.RequestID,
		})
	}
	w.Flush()
}

// GET /api/v1/audit-logs/summary
func (h *Handler) Summary(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	s, err := h.repo.Summary(c.Request.Context(), cl.OrganizationID)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "audit log summary retrieved", s)
}

// RegisterRoutes attaches all audit-log routes to the given RouterGroup.
// Caller is responsible for applying auth + permission middleware before calling this.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.List)
	rg.GET("/summary", h.Summary)
	rg.GET("/export", h.Export)
	rg.GET("/:id", h.GetByID)
}
