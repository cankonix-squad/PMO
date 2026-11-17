package kpbu

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/shared/types"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service implements KPBU K1 business rules. Every query is tenant-scoped.
type Service struct {
	db          *gorm.DB
	log         *zap.Logger
	now         func() time.Time
	storageRoot string
	maxSize     int64
	notifier    Notifier
}

func NewService(db *gorm.DB, log *zap.Logger, storageRoot string, maxSize int64) *Service {
	if maxSize <= 0 {
		maxSize = 20 << 20
	}
	return &Service{db: db, log: log, now: time.Now, storageRoot: storageRoot, maxSize: maxSize}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func today(now time.Time) time.Time {
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// delay returns how many days planned_end has passed without actual_end.
func delay(plannedEnd, actualEnd *types.FlexTime, now time.Time) int {
	if plannedEnd == nil || plannedEnd.IsZero() {
		return 0
	}
	end := today(now)
	if actualEnd != nil && !actualEnd.IsZero() {
		end = today(actualEnd.Time)
	}
	pe := today(plannedEnd.Time)
	if !end.After(pe) {
		return 0
	}
	return int(end.Sub(pe).Hours() / 24)
}

func (s *Service) existsInOrg(table string, orgID, id uuid.UUID) (bool, error) {
	var n int64
	err := s.db.Table(table).
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, orgID).
		Count(&n).Error
	return n > 0, err
}

func (s *Service) validateProject(orgID uuid.UUID, req *ProjectRequest) error {
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	req.Sector = strings.ToUpper(strings.TrimSpace(req.Sector))
	req.Initiative = strings.ToUpper(strings.TrimSpace(req.Initiative))
	req.ReturnScheme = strings.ToUpper(strings.TrimSpace(req.ReturnScheme))
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	if req.Initiative == "" {
		req.Initiative = "SOLICITED"
	}
	if req.ReturnScheme == "" {
		req.ReturnScheme = "AVAILABILITY_PAYMENT"
	}
	if req.Currency == "" {
		req.Currency = "IDR"
	}
	if req.Code == "" || req.Name == "" {
		return invalid("code and name are required")
	}
	if !validSectors[req.Sector] {
		return invalid("invalid sector")
	}
	if !validInitiatives[req.Initiative] {
		return invalid("invalid initiative")
	}
	if !validReturnSchemes[req.ReturnScheme] {
		return invalid("invalid return_scheme")
	}
	if req.ProcurementMethod != nil {
		m := strings.ToUpper(strings.TrimSpace(*req.ProcurementMethod))
		if m == "" {
			req.ProcurementMethod = nil
		} else if !validProcMethod[m] {
			return invalid("invalid procurement_method")
		} else {
			req.ProcurementMethod = &m
		}
	}
	if req.ProcurementMethod != nil && *req.ProcurementMethod == "PENUNJUKAN_LANGSUNG" && strings.TrimSpace(req.ProcurementReason) == "" {
		return invalid("procurement_reason is required for direct appointment (Perpres 38/2015 Ps.39)")
	}
	if (req.Latitude == nil) != (req.Longitude == nil) {
		return invalid("latitude and longitude must be provided together")
	}
	if req.ConcessionStart != nil && req.ConcessionEnd != nil &&
		!req.ConcessionStart.IsZero() && !req.ConcessionEnd.IsZero() &&
		req.ConcessionEnd.Before(req.ConcessionStart.Time) {
		return invalid("concession_end must not be before concession_start")
	}
	refs := []struct {
		table string
		id    *uuid.UUID
		msg   string
	}{
		{"projects", req.ProjectID, "project_id not found"},
		{"org_units", req.OrgUnitID, "org_unit_id not found"},
	}
	for _, r := range refs {
		if r.id == nil {
			continue
		}
		ok, err := s.existsInOrg(r.table, orgID, *r.id)
		if err != nil {
			return err
		}
		if !ok {
			return invalid(r.msg)
		}
	}
	return nil
}

func applyProject(p *Project, req ProjectRequest) {
	p.Code = req.Code
	p.Name = req.Name
	p.Description = req.Description
	p.Sector = req.Sector
	p.Initiative = req.Initiative
	p.ReturnScheme = req.ReturnScheme
	p.PJPKName = strings.TrimSpace(req.PJPKName)
	p.PJPKUnit = strings.TrimSpace(req.PJPKUnit)
	p.InvestmentValue = req.InvestmentValue
	p.Currency = req.Currency
	p.ConcessionYears = req.ConcessionYears
	p.ConcessionStart = req.ConcessionStart
	p.ConcessionEnd = req.ConcessionEnd
	p.FinancialCloseTarget = req.FinancialCloseTarget
	p.CODTarget = req.CODTarget
	p.ProcurementMethod = req.ProcurementMethod
	p.ProcurementReason = strings.TrimSpace(req.ProcurementReason)
	p.AgreementSignedDate = req.AgreementSignedDate
	p.ProjectID = req.ProjectID
	p.OrgUnitID = req.OrgUnitID
	p.Province = strings.TrimSpace(req.Province)
	p.Latitude = req.Latitude
	p.Longitude = req.Longitude
}

// enrich fills computed fields: org unit name and current-stage delay.
func (s *Service) enrich(orgID uuid.UUID, list []Project) error {
	if len(list) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(list))
	unitIDs := []uuid.UUID{}
	for _, p := range list {
		ids = append(ids, p.ID)
		if p.OrgUnitID != nil {
			unitIDs = append(unitIDs, *p.OrgUnitID)
		}
	}
	var ms []StageMilestone
	if err := s.db.Where("organization_id = ? AND kpbu_project_id IN ?", orgID, ids).Find(&ms).Error; err != nil {
		return err
	}
	byKey := map[string]StageMilestone{}
	for _, m := range ms {
		byKey[m.KPBUProjectID.String()+"|"+m.Stage] = m
	}
	names := map[uuid.UUID]string{}
	if len(unitIDs) > 0 {
		var rows []struct {
			ID   uuid.UUID
			Name string
		}
		if err := s.db.Table("org_units").Select("id, name").
			Where("id IN ? AND organization_id = ? AND deleted_at IS NULL", unitIDs, orgID).
			Scan(&rows).Error; err != nil {
			return err
		}
		for _, r := range rows {
			names[r.ID] = r.Name
		}
	}
	now := s.now()
	for i := range list {
		p := &list[i]
		if p.OrgUnitID != nil {
			p.OrgUnitName = names[*p.OrgUnitID]
		}
		if m, ok := byKey[p.ID.String()+"|"+p.Stage]; ok {
			p.StagePlannedEnd = m.PlannedEnd
			p.StageDelayDays = delay(m.PlannedEnd, m.ActualEnd, now)
			p.StageIsDelayed = p.StageDelayDays > 0
		}
		financialCloseStatus(p, now)
	}
	return nil
}

// List returns KPBU projects of the organization.
func (s *Service) List(orgID uuid.UUID, f ListFilter) ([]Project, error) {
	q := s.db.Where("organization_id = ?", orgID)
	if f.Stage != "" {
		q = q.Where("stage = ?", strings.ToUpper(f.Stage))
	}
	if f.Sector != "" {
		q = q.Where("sector = ?", strings.ToUpper(f.Sector))
	}
	if t := strings.TrimSpace(f.Search); t != "" {
		like := "%" + t + "%"
		q = q.Where("(code ILIKE ? OR name ILIKE ? OR pjpk_name ILIKE ?)", like, like, like)
	}
	var list []Project
	if err := q.Order("code ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	if err := s.enrich(orgID, list); err != nil {
		return nil, err
	}
	return list, nil
}

func (s *Service) findProject(orgID, id uuid.UUID) (*Project, error) {
	var p Project
	err := s.db.Where("id = ? AND organization_id = ?", id, orgID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, err
}

// Get returns a project with stage milestones, parties, and allowed next stages.
func (s *Service) Get(orgID, id uuid.UUID) (*Detail, error) {
	p, err := s.findProject(orgID, id)
	if err != nil {
		return nil, err
	}
	one := []Project{*p}
	if err := s.enrich(orgID, one); err != nil {
		return nil, err
	}
	*p = one[0]
	ms, err := s.ListStageMilestones(orgID, id)
	if err != nil {
		return nil, err
	}
	parties, err := s.ListParties(orgID, id)
	if err != nil {
		return nil, err
	}
	return &Detail{
		Project:         p,
		StageMilestones: ms,
		Parties:         parties,
		AllowedNext:     AllowedTransitions(p.Stage, p.PreviousStage),
	}, nil
}

// Create registers a new KPBU project at stage IDENTIFIED.
func (s *Service) Create(orgID, actorID uuid.UUID, req ProjectRequest) (*Project, error) {
	if err := s.validateProject(orgID, &req); err != nil {
		return nil, err
	}
	p := Project{OrganizationID: orgID, Stage: StageIdentified, CreatedBy: &actorID}
	applyProject(&p, req)
	if err := s.db.Create(&p).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, ErrCodeTaken
		}
		return nil, err
	}
	return &p, nil
}

// Update replaces the editable fields. Stage changes go through Transition.
func (s *Service) Update(orgID, id uuid.UUID, req ProjectRequest) (*Project, error) {
	p, err := s.findProject(orgID, id)
	if err != nil {
		return nil, err
	}
	if err := s.validateProject(orgID, &req); err != nil {
		return nil, err
	}
	applyProject(p, req)
	if err := s.db.Save(p).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, ErrCodeTaken
		}
		return nil, err
	}
	return p, nil
}

// Delete soft-deletes a project with its milestones and parties.
func (s *Service) Delete(orgID, id uuid.UUID) (*Project, error) {
	p, err := s.findProject(orgID, id)
	if err != nil {
		return nil, err
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("organization_id = ? AND kpbu_project_id = ?", orgID, id).Delete(&StageMilestone{}).Error; err != nil {
			return err
		}
		if err := tx.Where("organization_id = ? AND kpbu_project_id = ?", orgID, id).Delete(&Party{}).Error; err != nil {
			return err
		}
		return tx.Delete(p).Error
	})
	return p, err
}

// Transition moves a project along the stage FSM. It records the actual end of
// the stage being left and the actual start of the stage being entered, and sets
// financial_close_actual / cod_actual when those stages are reached.
func (s *Service) Transition(orgID, id uuid.UUID, req TransitionRequest) (*Project, string, error) {
	to := strings.ToUpper(strings.TrimSpace(req.ToStage))
	if !IsKnownStage(to) {
		return nil, "", ErrInvalidStage
	}
	p, err := s.findProject(orgID, id)
	if err != nil {
		return nil, "", err
	}
	from := p.Stage
	if !CanTransition(from, p.PreviousStage, to) {
		return nil, from, ErrInvalidTransition
	}
	date := types.FlexTime{Time: today(s.now())}
	if req.Date != nil && !req.Date.IsZero() {
		date = *req.Date
	}
	if req.OverrideDocumentCheck {
		if strings.TrimSpace(req.OverrideReason) == "" {
			return nil, from, invalid("override_reason is required when overriding the document check")
		}
	} else if missing, err := s.missingForStage(orgID, p, to); err != nil {
		return nil, from, err
	} else if len(missing) > 0 {
		return nil, from, &GateError{Stage: to, Missing: missing}
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if isMainStage(from) && to != StageOnHold {
			if err := s.touchMilestone(tx, orgID, id, from, nil, &date, req.Notes); err != nil {
				return err
			}
		}
		if isMainStage(to) && from != StageOnHold {
			if err := s.touchMilestone(tx, orgID, id, to, &date, nil, ""); err != nil {
				return err
			}
		}
		switch {
		case to == StageOnHold:
			prev := from
			p.PreviousStage = &prev
		case from == StageOnHold:
			p.PreviousStage = nil
		}
		if to == StageContractSigned && (p.AgreementSignedDate == nil || p.AgreementSignedDate.IsZero()) {
			d := date
			p.AgreementSignedDate = &d
		}
		if to == StageFinancialClose {
			d := date
			p.FinancialCloseActual = &d
		}
		if to == StageOperation {
			d := date
			p.CODActual = &d
		}
		p.Stage = to
		return tx.Save(p).Error
	})
	if err != nil {
		return nil, from, err
	}
	return p, from, nil
}

// touchMilestone sets actual start/end on a stage milestone, creating it when missing.
// Existing actual dates are kept.
func (s *Service) touchMilestone(tx *gorm.DB, orgID, projectID uuid.UUID, stage string, start, end *types.FlexTime, notes string) error {
	var m StageMilestone
	err := tx.Where("organization_id = ? AND kpbu_project_id = ? AND stage = ?", orgID, projectID, stage).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		m = StageMilestone{OrganizationID: orgID, KPBUProjectID: projectID, Stage: stage}
	} else if err != nil {
		return err
	}
	if start != nil && (m.ActualStart == nil || m.ActualStart.IsZero()) {
		d := *start
		m.ActualStart = &d
	}
	if end != nil && (m.ActualEnd == nil || m.ActualEnd.IsZero()) {
		d := *end
		m.ActualEnd = &d
	}
	// Append the transition note instead of replacing notes the user already wrote.
	if notes != "" {
		if m.Notes == "" {
			m.Notes = notes
		} else {
			m.Notes = m.Notes + "\n" + notes
		}
	}
	return tx.Save(&m).Error
}

// ListStageMilestones returns one row per main stage, in lifecycle order.
// Stages without a stored row are returned as empty placeholders (ID = nil UUID).
func (s *Service) ListStageMilestones(orgID, projectID uuid.UUID) ([]StageMilestone, error) {
	var stored []StageMilestone
	if err := s.db.Where("organization_id = ? AND kpbu_project_id = ?", orgID, projectID).Find(&stored).Error; err != nil {
		return nil, err
	}
	byStage := map[string]StageMilestone{}
	for _, m := range stored {
		byStage[m.Stage] = m
	}
	now := s.now()
	out := make([]StageMilestone, 0, len(MainStages))
	for _, st := range MainStages {
		m, ok := byStage[st]
		if !ok {
			m = StageMilestone{OrganizationID: orgID, KPBUProjectID: projectID, Stage: st}
		}
		m.DelayDays = delay(m.PlannedEnd, m.ActualEnd, now)
		m.IsDelayed = m.DelayDays > 0
		out = append(out, m)
	}
	return out, nil
}

// UpsertStageMilestone sets planned/actual dates of one stage.
func (s *Service) UpsertStageMilestone(orgID, projectID uuid.UUID, req StageMilestoneRequest) (*StageMilestone, error) {
	stage := strings.ToUpper(strings.TrimSpace(req.Stage))
	if !isMainStage(stage) {
		return nil, invalid("stage must be one of the main lifecycle stages")
	}
	if _, err := s.findProject(orgID, projectID); err != nil {
		return nil, err
	}
	for _, pair := range [][2]*types.FlexTime{{req.PlannedStart, req.PlannedEnd}, {req.ActualStart, req.ActualEnd}} {
		if pair[0] != nil && pair[1] != nil && !pair[0].IsZero() && !pair[1].IsZero() && pair[1].Before(pair[0].Time) {
			return nil, invalid("end date must not be before start date")
		}
	}
	var m StageMilestone
	err := s.db.Where("organization_id = ? AND kpbu_project_id = ? AND stage = ?", orgID, projectID, stage).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		m = StageMilestone{OrganizationID: orgID, KPBUProjectID: projectID, Stage: stage}
	} else if err != nil {
		return nil, err
	}
	m.PlannedStart, m.PlannedEnd = req.PlannedStart, req.PlannedEnd
	m.ActualStart, m.ActualEnd = req.ActualStart, req.ActualEnd
	m.Notes = req.Notes
	if err := s.db.Save(&m).Error; err != nil {
		return nil, err
	}
	m.DelayDays = delay(m.PlannedEnd, m.ActualEnd, s.now())
	m.IsDelayed = m.DelayDays > 0
	return &m, nil
}

// ListParties returns parties of a project.
func (s *Service) ListParties(orgID, projectID uuid.UUID) ([]Party, error) {
	var list []Party
	err := s.db.Where("organization_id = ? AND kpbu_project_id = ?", orgID, projectID).
		Order("role ASC, name ASC").Find(&list).Error
	return list, err
}

func (s *Service) validateParty(orgID uuid.UUID, req *PartyRequest) error {
	req.Role = strings.ToUpper(strings.TrimSpace(req.Role))
	req.Name = strings.TrimSpace(req.Name)
	if !validPartyRoles[req.Role] {
		return invalid("invalid party role")
	}
	if req.Name == "" {
		return invalid("name is required")
	}
	if req.VendorID != nil {
		ok, err := s.existsInOrg("vendors", orgID, *req.VendorID)
		if err != nil {
			return err
		}
		if !ok {
			return invalid("vendor_id not found")
		}
	}
	return nil
}

func applyParty(p *Party, req PartyRequest) {
	p.Role, p.Name, p.VendorID, p.SharePct = req.Role, req.Name, req.VendorID, req.SharePct
	p.ContactPerson = strings.TrimSpace(req.ContactPerson)
	p.Email = strings.TrimSpace(req.Email)
	p.Phone = strings.TrimSpace(req.Phone)
	p.Notes = req.Notes
}

// CreateParty adds a party to a project.
func (s *Service) CreateParty(orgID, projectID uuid.UUID, req PartyRequest) (*Party, error) {
	if _, err := s.findProject(orgID, projectID); err != nil {
		return nil, err
	}
	if err := s.validateParty(orgID, &req); err != nil {
		return nil, err
	}
	p := Party{OrganizationID: orgID, KPBUProjectID: projectID}
	applyParty(&p, req)
	return &p, s.db.Create(&p).Error
}

func (s *Service) findParty(orgID, projectID, partyID uuid.UUID) (*Party, error) {
	var p Party
	err := s.db.Where("id = ? AND organization_id = ? AND kpbu_project_id = ?", partyID, orgID, projectID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, err
}

// UpdateParty replaces a party's fields.
func (s *Service) UpdateParty(orgID, projectID, partyID uuid.UUID, req PartyRequest) (*Party, error) {
	p, err := s.findParty(orgID, projectID, partyID)
	if err != nil {
		return nil, err
	}
	if err := s.validateParty(orgID, &req); err != nil {
		return nil, err
	}
	applyParty(p, req)
	return p, s.db.Save(p).Error
}

// DeleteParty soft-deletes a party.
func (s *Service) DeleteParty(orgID, projectID, partyID uuid.UUID) (*Party, error) {
	p, err := s.findParty(orgID, projectID, partyID)
	if err != nil {
		return nil, err
	}
	return p, s.db.Delete(p).Error
}
