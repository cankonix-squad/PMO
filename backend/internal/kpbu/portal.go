package kpbu

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	notification "github.com/harmanto-49/cankora/internal/notifications/delivery"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"gorm.io/gorm"
)

// BUP portal: an external account (users.vendor_id set) may access a KPBU
// project when that vendor is recorded as the project's BUP party. Accounts
// are the same contractor accounts created at /settings/vendor-accounts.

var ErrNotPortalUser = errors.New("kpbu: account is not linked to a business entity")

const ctxPortal = "kpbu_portal_vendor"

// PortalProject is the BUP-facing project summary.
type PortalProject struct {
	ID              uuid.UUID `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Stage           string    `json:"stage"`
	Sector          string    `json:"sector"`
	ReturnScheme    string    `json:"return_scheme"`
	PJPKName        string    `json:"pjpk_name"`
	InvestmentValue float64   `json:"investment_value"`
	Currency        string    `json:"currency"`
	PartyName       string    `json:"party_name"`
}

func (s *Service) vendorOf(orgID, userID uuid.UUID) (uuid.UUID, error) {
	var row struct{ VendorID *uuid.UUID }
	err := s.db.Table("users").Select("vendor_id").
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL AND is_active", userID, orgID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && row.VendorID == nil) {
		return uuid.Nil, ErrNotPortalUser
	}
	return derefID(row.VendorID), err
}

func derefID(p *uuid.UUID) uuid.UUID {
	if p == nil {
		return uuid.Nil
	}
	return *p
}

func (s *Service) portalQuery(orgID, vendorID uuid.UUID) *gorm.DB {
	return s.db.Table("kpbu_projects p").
		Select("p.id, p.code, p.name, p.stage, p.sector, p.return_scheme, p.pjpk_name, p.investment_value, p.currency, pa.name AS party_name").
		Joins("JOIN kpbu_parties pa ON pa.kpbu_project_id = p.id AND pa.deleted_at IS NULL AND pa.role = 'BUP' AND pa.vendor_id = ?", vendorID).
		Where("p.organization_id = ? AND p.deleted_at IS NULL", orgID)
}

func (s *Service) PortalProjects(orgID, vendorID uuid.UUID) ([]PortalProject, error) {
	var list []PortalProject
	return list, s.portalQuery(orgID, vendorID).Order("p.code ASC").Scan(&list).Error
}

func (s *Service) PortalProject(orgID, vendorID, projectID uuid.UUID) (*PortalProject, error) {
	var list []PortalProject
	if err := s.portalQuery(orgID, vendorID).Where("p.id = ?", projectID).Limit(1).Scan(&list).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, ErrNotFound
	}
	return &list[0], nil
}

// ── Handlers ────────────────────────────────────────────────────────────────

func (h *Handler) portalVendor(c *gin.Context) (uuid.UUID, bool) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return uuid.Nil, false
	}
	vid, err := h.svc.vendorOf(claims.OrganizationID, claims.UserID)
	if errors.Is(err, ErrNotPortalUser) {
		response.Forbidden(c, "this account is not linked to a business entity")
		return uuid.Nil, false
	}
	if err != nil {
		h.writeErr(c, err, "")
		return uuid.Nil, false
	}
	return vid, true
}

// PortalScope ensures the account is the BUP of project :id. Handlers after
// it can rely on the project being accessible to this business entity.
func (h *Handler) PortalScope(c *gin.Context) {
	vid, ok := h.portalVendor(c)
	if !ok {
		c.Abort()
		return
	}
	pid, ok := parseID(c, "id")
	if !ok {
		c.Abort()
		return
	}
	claims := claimsFromGin(c)
	if _, err := h.svc.PortalProject(claims.OrganizationID, vid, pid); err != nil {
		h.writeErr(c, err, "KPBU project not found")
		c.Abort()
		return
	}
	c.Set(ctxPortal, vid)
	c.Next()
}

func isPortal(c *gin.Context) bool {
	_, ok := c.Get(ctxPortal)
	return ok
}

// PortalListProjects GET /vendor-portal/kpbu-projects
func (h *Handler) PortalListProjects(c *gin.Context) {
	vid, ok := h.portalVendor(c)
	if !ok {
		return
	}
	list, err := h.svc.PortalProjects(claimsFromGin(c).OrganizationID, vid)
	if err != nil {
		h.writeErr(c, err, "")
		return
	}
	response.OK(c, "KPBU projects retrieved", list)
}

// PortalGetProject GET /vendor-portal/kpbu-projects/:id (after PortalScope)
func (h *Handler) PortalGetProject(c *gin.Context) {
	claims := claimsFromGin(c)
	vid, _ := c.Get(ctxPortal)
	pid, _ := parseID(c, "id")
	p, err := h.svc.PortalProject(claims.OrganizationID, vid.(uuid.UUID), pid)
	if err != nil {
		h.writeErr(c, err, "KPBU project not found")
		return
	}
	response.OK(c, "KPBU project retrieved", p)
}

// ── KPI handlers (internal and portal) ──────────────────────────────────────

func (h *Handler) KPIRoutes() kpiHandlers {
	return kpiHandlers{h: h}
}

type kpiHandlers struct{ h *Handler }

func (k kpiHandlers) List(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	pid, ok := parseID(c, "id")
	if !ok {
		return
	}
	list, err := k.h.svc.ListKPIs(claims.OrganizationID, pid, isPortal(c) || c.Query("active") == "true")
	if err != nil {
		k.h.writeErr(c, err, "KPBU project not found")
		return
	}
	response.OK(c, "KPBU KPIs retrieved", list)
}

func (k kpiHandlers) save(c *gin.Context, withID bool) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	pid, ok := parseID(c, "id")
	if !ok {
		return
	}
	var id *uuid.UUID
	if withID {
		kid, ok := parseID(c, "kpiID")
		if !ok {
			return
		}
		id = &kid
	}
	var req KPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	v, err := k.h.svc.SaveKPI(claims.OrganizationID, pid, id, req)
	if err != nil {
		k.h.writeErr(c, err, "KPBU KPI not found")
		return
	}
	if withID {
		k.h.record(c, claims, "kpbu_kpi.updated", "kpbu_kpi", v.ID, v.Code, v)
		response.OK(c, "KPBU KPI updated", v)
		return
	}
	k.h.record(c, claims, "kpbu_kpi.created", "kpbu_kpi", v.ID, v.Code, v)
	response.Created(c, "KPBU KPI created", v)
}

func (k kpiHandlers) Create(c *gin.Context) { k.save(c, false) }
func (k kpiHandlers) Update(c *gin.Context) { k.save(c, true) }

func (k kpiHandlers) Delete(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	pid, ok := parseID(c, "id")
	if !ok {
		return
	}
	kid, ok := parseID(c, "kpiID")
	if !ok {
		return
	}
	v, err := k.h.svc.DeleteKPI(claims.OrganizationID, pid, kid)
	if err != nil {
		k.h.writeErr(c, err, "KPBU KPI not found")
		return
	}
	k.h.record(c, claims, "kpbu_kpi.deleted", "kpbu_kpi", v.ID, v.Code, nil)
	response.NoContent(c)
}

func queryInt(c *gin.Context, name string) int {
	v, _ := strconv.Atoi(c.Query(name))
	return v
}

func (k kpiHandlers) ListMeasurements(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	pid, ok := parseID(c, "id")
	if !ok {
		return
	}
	list, err := k.h.svc.ListMeasurements(claims.OrganizationID, pid, c.Query("status"), queryInt(c, "year"), queryInt(c, "month"))
	if err != nil {
		k.h.writeErr(c, err, "KPBU project not found")
		return
	}
	response.OK(c, "KPBU KPI measurements retrieved", list)
}

func (k kpiHandlers) CreateMeasurement(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	pid, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req MeasurementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	source := MeasureInternal
	if isPortal(c) {
		source = MeasurePortal
	}
	m, err := k.h.svc.CreateMeasurement(claims.OrganizationID, claims.UserID, pid, source, req)
	if err != nil {
		k.h.writeErr(c, err, "KPBU project not found")
		return
	}
	k.h.record(c, claims, "kpbu_kpi_measurement.created", "kpbu_kpi_measurement", m.ID, m.KPICode, m)
	if source == MeasurePortal {
		k.h.svc.notify(claims.OrganizationID, k.h.svc.internalReviewers(claims.OrganizationID), notification.PriorityNormal,
			"Laporan KPI layanan KPBU menunggu verifikasi",
			fmt.Sprintf("Badan usaha melaporkan KPI %s (%s) periode %s: %v.", m.KPICode, m.KPIName, monthYear(m.PeriodYear, m.PeriodMonth), m.ActualValue),
			"kpbu_kpi_measurement", m.ID)
	}
	response.Created(c, "KPBU KPI measurement recorded", m)
}

func (k kpiHandlers) ReviewMeasurement(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	pid, ok := parseID(c, "id")
	if !ok {
		return
	}
	mid, ok := parseID(c, "measurementID")
	if !ok {
		return
	}
	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	m, err := k.h.svc.ReviewMeasurement(claims.OrganizationID, claims.UserID, pid, mid, req)
	if err != nil {
		k.h.writeErr(c, err, "KPBU KPI measurement not found")
		return
	}
	k.h.record(c, claims, "kpbu_kpi_measurement.reviewed", "kpbu_kpi_measurement", m.ID, m.KPICode+" "+m.Status, m)
	if m.Source == MeasurePortal && m.SubmittedBy != nil {
		subject, prio := "Laporan KPI diverifikasi", notification.PriorityNormal
		body := fmt.Sprintf("KPI %s periode %s telah diverifikasi.", m.KPICode, monthYear(m.PeriodYear, m.PeriodMonth))
		if m.Status == "REJECTED" {
			subject, prio = "Laporan KPI ditolak", notification.PriorityHigh
			body = fmt.Sprintf("KPI %s periode %s ditolak. Alasan: %s. Kirim ulang nilai yang benar.", m.KPICode, monthYear(m.PeriodYear, m.PeriodMonth), m.ReviewNotes)
		}
		k.h.svc.notify(claims.OrganizationID, []uuid.UUID{*m.SubmittedBy}, prio, subject, body, "kpbu_kpi_measurement", m.ID)
	}
	response.OK(c, "KPBU KPI measurement reviewed", m)
}

func (k kpiHandlers) Scorecard(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	pid, ok := parseID(c, "id")
	if !ok {
		return
	}
	sc, err := k.h.svc.KPIScorecard(claims.OrganizationID, pid, queryInt(c, "year"), queryInt(c, "month"))
	if err != nil {
		k.h.writeErr(c, err, "KPBU project not found")
		return
	}
	response.OK(c, "KPBU KPI scorecard retrieved", sc)
}

// SetNotifier on the handler forwards to the service.
func (h *Handler) SetNotifier(n Notifier) { h.svc.SetNotifier(n) }
