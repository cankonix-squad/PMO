package government

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
)

// Handler exposes government connector integration endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new government integration Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
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

// RegisterRoutes wires all government integration routes onto the provided RouterGroup.
// The group should already have AuthRequired middleware applied.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/connectors", h.ListConnectors)
	rg.GET("/connectors/:key", h.GetConnector)
	rg.GET("/config", h.GetConfig)
	rg.GET("/runs", h.ListRuns)
	rg.POST("/runs", h.CreateRun)
	rg.GET("/runs/:id", h.GetRun)
	rg.POST("/runs/:id/cancel", h.CancelRun)
	rg.GET("/runs/:id/records", h.ListRecords)

	// Mappings — general listing
	rg.GET("/mappings", h.ListMappings)

	// P3-002: Entity Resolution endpoints
	// NOTE: /mappings/pending must be registered BEFORE /mappings/:id to avoid
	// Gin treating "pending" as the :id parameter.
	rg.GET("/mappings/pending", h.ListPendingMappings)
	rg.GET("/mappings/:id", h.GetMapping)
	rg.GET("/mappings/:id/candidates", h.GetMappingCandidates)
	rg.POST("/mappings/:id/match", h.MatchMapping)
	rg.POST("/mappings/:id/unmatch", h.UnmatchMapping)
	rg.POST("/mappings/:id/reject", h.RejectMapping)
}

// ---------------------------------------------------------------------------
// GET /api/v1/integrations/government/connectors
// ---------------------------------------------------------------------------

func (h *Handler) ListConnectors(c *gin.Context) {
	_, ok := h.claims(c)
	if !ok {
		return
	}
	connectors := h.svc.ListConnectors()
	response.OK(c, "connectors retrieved", gin.H{
		"data": connectors,
		"meta": gin.H{"total": len(connectors)},
	})
}

// ---------------------------------------------------------------------------
// GET /api/v1/integrations/government/connectors/:key
// ---------------------------------------------------------------------------

func (h *Handler) GetConnector(c *gin.Context) {
	_, ok := h.claims(c)
	if !ok {
		return
	}
	key := c.Param("key")
	connector, err := h.svc.GetConnector(key)
	if errors.Is(err, ErrInvalidConnector) {
		response.NotFound(c, "connector not found")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, "connector retrieved", connector)
}

// ---------------------------------------------------------------------------
// GET /api/v1/integrations/government/config
// ---------------------------------------------------------------------------

func (h *Handler) GetConfig(c *gin.Context) {
	_, ok := h.claims(c)
	if !ok {
		return
	}
	cfg := h.svc.GetConfig()
	response.OK(c, "connector config retrieved", gin.H{
		"data": cfg,
	})
}

// ---------------------------------------------------------------------------
// GET /api/v1/integrations/government/runs
// ---------------------------------------------------------------------------

func (h *Handler) ListRuns(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	f := ListRunsFilter{
		ConnectorKey: c.Query("connector_key"),
		DatasetType:  c.Query("dataset_type"),
		Status:       c.Query("status"),
	}
	f.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	f.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	orgID, err := uuid.Parse(cl.OrganizationID.String())
	if err != nil {
		response.BadRequest(c, "invalid organization")
		return
	}

	runs, total, err := h.svc.ListRuns(c.Request.Context(), orgID, f)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "sync runs retrieved", gin.H{
		"data": runs,
		"meta": gin.H{
			"total":     total,
			"page":      f.Page,
			"page_size": f.PageSize,
		},
	})
}

// ---------------------------------------------------------------------------
// POST /api/v1/integrations/government/runs
// ---------------------------------------------------------------------------

func (h *Handler) CreateRun(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	var req CreateRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	orgID := cl.OrganizationID
	userID := cl.UserID

	run, err := h.svc.CreateRun(c.Request.Context(), orgID, userID, req)
	if errors.Is(err, ErrInvalidConnector) {
		response.BadRequest(c, "invalid connector_key")
		return
	}
	if errors.Is(err, ErrInvalidDatasetType) {
		response.BadRequest(c, "dataset_type not supported by this connector")
		return
	}
	if errors.Is(err, ErrInvalidMode) {
		response.BadRequest(c, "mode must be SAMPLE, DRY_RUN, or COMMIT")
		return
	}
	if errors.Is(err, ErrDuplicateRun) {
		// Return the existing run rather than an error — idempotent behaviour
		response.OK(c, "existing run returned (idempotent)", run)
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}

	// Auto-process the run immediately
	processed, processErr := h.svc.ProcessRun(c.Request.Context(), orgID, run.ID)
	if processErr != nil {
		// Run was created but processing failed; return the pending run with a warning
		response.Created(c, "sync run created (processing failed)", run)
		return
	}

	response.Created(c, "sync run created and processed", processed)
}

// ---------------------------------------------------------------------------
// GET /api/v1/integrations/government/runs/:id
// ---------------------------------------------------------------------------

func (h *Handler) GetRun(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid run id")
		return
	}

	run, err := h.svc.GetRun(c.Request.Context(), cl.OrganizationID, runID)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "sync run not found")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "sync run retrieved", run)
}

// ---------------------------------------------------------------------------
// POST /api/v1/integrations/government/runs/:id/cancel
// ---------------------------------------------------------------------------

func (h *Handler) CancelRun(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid run id")
		return
	}

	run, err := h.svc.CancelRun(c.Request.Context(), cl.OrganizationID, runID, cl.UserID)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "sync run not found")
		return
	}
	if errors.Is(err, ErrInvalidTransition) {
		response.BadRequest(c, "only PENDING runs can be cancelled")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "sync run cancelled", run)
}

// ---------------------------------------------------------------------------
// GET /api/v1/integrations/government/runs/:id/records
// ---------------------------------------------------------------------------

func (h *Handler) ListRecords(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid run id")
		return
	}

	f := ListRecordsFilter{
		Status: c.Query("status"),
		Action: c.Query("action"),
	}
	f.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	f.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	records, total, err := h.svc.ListRecords(c.Request.Context(), cl.OrganizationID, runID, f)
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, "sync run not found")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "sync records retrieved", gin.H{
		"data": records,
		"meta": gin.H{
			"total":     total,
			"page":      f.Page,
			"page_size": f.PageSize,
		},
	})
}

// ---------------------------------------------------------------------------
// GET /api/v1/integrations/government/mappings
// ---------------------------------------------------------------------------

func (h *Handler) ListMappings(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	f := ListMappingsFilter{
		ConnectorKey:       c.Query("connector_key"),
		DatasetType:        c.Query("dataset_type"),
		InternalEntityType: c.Query("internal_entity_type"),
		MatchStatus:        c.Query("match_status"),
	}
	f.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	f.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	mappings, total, err := h.svc.ListMappings(c.Request.Context(), cl.OrganizationID, f)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "external mappings retrieved", gin.H{
		"data": mappings,
		"meta": gin.H{
			"total":     total,
			"page":      f.Page,
			"page_size": f.PageSize,
		},
	})
}

// ---------------------------------------------------------------------------
// GET /api/v1/integrations/government/mappings/pending
// ---------------------------------------------------------------------------

func (h *Handler) ListPendingMappings(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	f := ListPendingMappingsFilter{
		ConnectorKey: c.Query("connector_key"),
		DatasetType:  c.Query("dataset_type"),
	}
	f.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	f.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	mappings, total, err := h.svc.ListPendingMappings(c.Request.Context(), cl.OrganizationID, f)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "pending mappings retrieved", gin.H{
		"data": mappings,
		"meta": gin.H{
			"total":     total,
			"page":      f.Page,
			"page_size": f.PageSize,
		},
	})
}

// ---------------------------------------------------------------------------
// GET /api/v1/integrations/government/mappings/:id
// ---------------------------------------------------------------------------

func (h *Handler) GetMapping(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	mappingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid mapping id")
		return
	}

	mapping, err := h.svc.GetMapping(c.Request.Context(), cl.OrganizationID, mappingID)
	if errors.Is(err, ErrMappingNotFound) {
		response.NotFound(c, "mapping not found")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "mapping retrieved", mapping)
}

// ---------------------------------------------------------------------------
// GET /api/v1/integrations/government/mappings/:id/candidates
// ---------------------------------------------------------------------------

func (h *Handler) GetMappingCandidates(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	mappingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid mapping id")
		return
	}

	candidates, err := h.svc.GetCandidates(c.Request.Context(), cl.OrganizationID, mappingID)
	if errors.Is(err, ErrMappingNotFound) {
		response.NotFound(c, "mapping not found")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "candidates retrieved", gin.H{
		"data": candidates,
		"meta": gin.H{"total": len(candidates)},
	})
}

// ---------------------------------------------------------------------------
// POST /api/v1/integrations/government/mappings/:id/match
// ---------------------------------------------------------------------------

func (h *Handler) MatchMapping(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	mappingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid mapping id")
		return
	}

	var req MatchMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	mapping, err := h.svc.MatchMapping(c.Request.Context(), cl.OrganizationID, cl.UserID, mappingID, req)
	if errors.Is(err, ErrMappingNotFound) {
		response.NotFound(c, "mapping not found")
		return
	}
	if errors.Is(err, ErrAlreadyMatched) {
		response.BadRequest(c, "mapping is already matched")
		return
	}
	if errors.Is(err, ErrEntityTypeMismatch) {
		response.BadRequest(c, "internal_entity_type does not match mapping dataset_type")
		return
	}
	if errors.Is(err, ErrCandidateNotFound) {
		response.NotFound(c, "target entity not found in this organisation")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "mapping matched", mapping)
}

// ---------------------------------------------------------------------------
// POST /api/v1/integrations/government/mappings/:id/unmatch
// ---------------------------------------------------------------------------

func (h *Handler) UnmatchMapping(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	mappingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid mapping id")
		return
	}

	mapping, err := h.svc.UnmatchMapping(c.Request.Context(), cl.OrganizationID, cl.UserID, mappingID)
	if errors.Is(err, ErrMappingNotFound) {
		response.NotFound(c, "mapping not found")
		return
	}
	if errors.Is(err, ErrInvalidTransition) {
		response.BadRequest(c, "only MATCHED mappings can be unmatched")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "mapping unmatched", mapping)
}

// ---------------------------------------------------------------------------
// POST /api/v1/integrations/government/mappings/:id/reject
// ---------------------------------------------------------------------------

func (h *Handler) RejectMapping(c *gin.Context) {
	cl, ok := h.claims(c)
	if !ok {
		return
	}

	mappingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid mapping id")
		return
	}

	var req RejectMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	mapping, err := h.svc.RejectMapping(c.Request.Context(), cl.OrganizationID, cl.UserID, mappingID, req)
	if errors.Is(err, ErrMappingNotFound) {
		response.NotFound(c, "mapping not found")
		return
	}
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "mapping rejected", mapping)
}
