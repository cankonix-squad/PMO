package priority

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/audit"
	"gorm.io/gorm"
)

var (
	ErrNotFound         = errors.New("priority resource not found")
	ErrActiveExists     = errors.New("an ACTIVE formula already exists for this organization")
	ErrFormulaNotDraft  = errors.New("only DRAFT formulas can be activated")
	ErrFormulaNotActive = errors.New("only ACTIVE formulas can be archived")
	ErrWeightSumInvalid = errors.New("component weights must sum to 1.0 (±0.01)")
	ErrNoActiveFormula  = errors.New("no ACTIVE priority formula found for this organization")
)

// Service handles all priority scoring logic.
type Service struct {
	db    *gorm.DB
	audit *audit.Writer
}

func NewService(db *gorm.DB, a *audit.Writer) *Service { return &Service{db: db, audit: a} }

// ------------------------------------------------------------------
// Formula CRUD
// ------------------------------------------------------------------

func (s *Service) ListFormulas(org uuid.UUID) ([]Formula, error) {
	var v []Formula
	err := s.db.Where("organization_id = ? AND deleted_at IS NULL", org).
		Preload("Components").
		Order("version DESC").Find(&v).Error
	return v, err
}

func (s *Service) GetFormula(org, id uuid.UUID) (*Formula, error) {
	var f Formula
	err := s.db.Where("organization_id = ? AND id = ? AND deleted_at IS NULL", org, id).
		Preload("Components").First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &f, err
}

func (s *Service) CreateFormula(org, actor uuid.UUID, r CreateFormulaRequest) (*Formula, error) {
	if err := validateWeights(r.Components); err != nil {
		return nil, err
	}
	if r.MissingDataRule == "" {
		r.MissingDataRule = MissingPenalize
	}
	thresholds := defaultThresholds()
	if r.CategoryThresholds != nil {
		thresholds = r.CategoryThresholds
	}
	tb, _ := json.Marshal(thresholds)

	now := time.Now()
	f := &Formula{
		ID:                 uuid.New(),
		OrganizationID:     org,
		Name:               r.Name,
		Description:        r.Description,
		Version:            1,
		Status:             FormulaStatusDraft,
		MissingDataRule:    r.MissingDataRule,
		CategoryThresholds: tb,
		CreatedBy:          &actor,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.db.Create(f).Error; err != nil {
		return nil, err
	}
	for i, c := range r.Components {
		comp := &FormulaComponent{
			ID:             uuid.New(),
			FormulaID:      f.ID,
			OrganizationID: org,
			ComponentKey:   c.ComponentKey,
			Label:          c.Label,
			Weight:         c.Weight,
			SortOrder:      c.SortOrder,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if comp.SortOrder == 0 {
			comp.SortOrder = i + 1
		}
		if err := s.db.Create(comp).Error; err != nil {
			return nil, err
		}
	}
	s.record(org, actor, "priority.formula.created", "priority_formula", f.ID.String())
	return s.GetFormula(org, f.ID)
}

func (s *Service) UpdateFormula(org, actor, id uuid.UUID, r UpdateFormulaRequest) (*Formula, error) {
	f, err := s.GetFormula(org, id)
	if err != nil {
		return nil, err
	}
	if f.Status != FormulaStatusDraft {
		return nil, errors.New("only DRAFT formulas can be updated")
	}
	now := time.Now()
	if r.Name != "" {
		f.Name = r.Name
	}
	if r.Description != "" {
		f.Description = r.Description
	}
	if r.MissingDataRule != "" {
		f.MissingDataRule = r.MissingDataRule
	}
	if r.CategoryThresholds != nil {
		tb, _ := json.Marshal(r.CategoryThresholds)
		f.CategoryThresholds = tb
	}
	f.UpdatedAt = now
	if err := s.db.Save(f).Error; err != nil {
		return nil, err
	}
	if len(r.Components) > 0 {
		if err := validateWeights(r.Components); err != nil {
			return nil, err
		}
		// Replace components
		if err := s.db.Where("formula_id = ?", id).Delete(&FormulaComponent{}).Error; err != nil {
			return nil, err
		}
		for i, c := range r.Components {
			comp := &FormulaComponent{
				ID:             uuid.New(),
				FormulaID:      f.ID,
				OrganizationID: org,
				ComponentKey:   c.ComponentKey,
				Label:          c.Label,
				Weight:         c.Weight,
				SortOrder:      c.SortOrder,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			if comp.SortOrder == 0 {
				comp.SortOrder = i + 1
			}
			if err := s.db.Create(comp).Error; err != nil {
				return nil, err
			}
		}
	}
	s.record(org, actor, "priority.formula.updated", "priority_formula", id.String())
	return s.GetFormula(org, id)
}

// ActivateFormula transitions a DRAFT formula to ACTIVE, archiving the previous ACTIVE one.
func (s *Service) ActivateFormula(org, actor, id uuid.UUID) (*Formula, error) {
	f, err := s.GetFormula(org, id)
	if err != nil {
		return nil, err
	}
	if f.Status != FormulaStatusDraft {
		return nil, ErrFormulaNotDraft
	}
	// Archive any currently ACTIVE formula
	now := time.Now()
	s.db.Model(&Formula{}).
		Where("organization_id = ? AND status = ? AND deleted_at IS NULL", org, FormulaStatusActive).
		Updates(map[string]interface{}{"status": FormulaStatusArchived, "updated_at": now})

	f.Status = FormulaStatusActive
	f.ActivatedBy = &actor
	f.ActivatedAt = &now
	f.UpdatedAt = now
	if err := s.db.Save(f).Error; err != nil {
		return nil, err
	}
	s.record(org, actor, "priority.formula.activated", "priority_formula", id.String())
	return s.GetFormula(org, id)
}

// ------------------------------------------------------------------
// Score calculation
// ------------------------------------------------------------------

// Calculate computes the priority score for a single project and persists the result.
func (s *Service) Calculate(org, actor, projectID uuid.UUID, r CalculateRequest) (*Score, error) {
	formula, err := s.resolveFormula(org, r.FormulaID)
	if err != nil {
		return nil, err
	}
	// Verify project belongs to org and is not deleted
	if err := s.verifyProject(org, projectID); err != nil {
		return nil, err
	}
	score, err := s.computeScore(org, actor, projectID, formula)
	if err != nil {
		return nil, err
	}
	s.record(org, actor, "priority.score.calculated", "project_priority_score", score.ID.String())
	return s.GetScore(org, projectID)
}

// BatchCalculate computes priority scores for all (or specified) active projects.
func (s *Service) BatchCalculate(org, actor uuid.UUID, r BatchCalculateRequest) (*BatchCalculateResponse, error) {
	formula, err := s.resolveFormula(org, r.FormulaID)
	if err != nil {
		return nil, err
	}
	// Build project list
	var projectIDs []uuid.UUID
	if len(r.ProjectIDs) > 0 {
		// Verify each project is in this org
		for _, pid := range r.ProjectIDs {
			if err := s.verifyProject(org, pid); err != nil {
				return nil, fmt.Errorf("project %s: %w", pid, err)
			}
		}
		projectIDs = r.ProjectIDs
	} else {
		// All active (non-deleted) projects in org
		type row struct{ ID uuid.UUID }
		var rows []row
		if err := s.db.Raw(
			"SELECT id FROM projects WHERE organization_id = ? AND deleted_at IS NULL",
			org,
		).Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, rr := range rows {
			projectIDs = append(projectIDs, rr.ID)
		}
	}

	calculated := 0
	skipped := 0
	var scores []*Score
	for _, pid := range projectIDs {
		sc, err := s.computeScore(org, actor, pid, formula)
		if err != nil {
			skipped++
			continue
		}
		scores = append(scores, sc)
		calculated++
	}

	// Assign ranks (highest total_score = rank 1)
	for i, sc := range scores {
		rank := i + 1
		s.db.Model(&Score{}).Where("id = ?", sc.ID).Update("rank_in_org", rank)
	}

	s.record(org, actor, "priority.batch.calculated", "priority_batch",
		fmt.Sprintf("formula=%s calculated=%d", formula.ID, calculated))
	return &BatchCalculateResponse{
		Calculated:     calculated,
		Skipped:        skipped,
		FormulaID:      formula.ID.String(),
		FormulaVersion: formula.Version,
	}, nil
}

// GetScore returns the latest score snapshot for a project.
func (s *Service) GetScore(org, projectID uuid.UUID) (*Score, error) {
	var sc Score
	err := s.db.Where("organization_id = ? AND project_id = ?", org, projectID).
		Preload("Components").
		Order("calculated_at DESC").First(&sc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &sc, err
}

// ListRanking returns the latest score per project, sorted by total_score DESC.
func (s *Service) ListRanking(org uuid.UUID, category, formulaID string) (*RankingResponse, error) {
	// Get the latest score per project using a subquery
	query := s.db.Raw(`
		SELECT DISTINCT ON (pps.project_id)
			pps.id          AS score_id,
			pps.project_id,
			p.name          AS project_name,
			p.status        AS project_status,
			pps.formula_id,
			pf.name         AS formula_name,
			pf.version      AS formula_version,
			pps.total_score,
			pps.score_category,
			pps.rank_in_org,
			pps.missing_components,
			pps.calculated_at
		FROM project_priority_scores pps
		JOIN projects p       ON p.id = pps.project_id AND p.deleted_at IS NULL
		JOIN priority_formulas pf ON pf.id = pps.formula_id AND pf.deleted_at IS NULL
		WHERE pps.organization_id = ?
		  AND p.organization_id   = ?
	`, org, org)

	if category != "" {
		query = s.db.Raw(`
			SELECT DISTINCT ON (pps.project_id)
				pps.id          AS score_id,
				pps.project_id,
				p.name          AS project_name,
				p.status        AS project_status,
				pps.formula_id,
				pf.name         AS formula_name,
				pf.version      AS formula_version,
				pps.total_score,
				pps.score_category,
				pps.rank_in_org,
				pps.missing_components,
				pps.calculated_at
			FROM project_priority_scores pps
			JOIN projects p       ON p.id = pps.project_id AND p.deleted_at IS NULL
			JOIN priority_formulas pf ON pf.id = pps.formula_id AND pf.deleted_at IS NULL
			WHERE pps.organization_id = ?
			  AND p.organization_id   = ?
			  AND pps.score_category  = ?
		`, org, org, category)
	}

	if formulaID != "" {
		fid, err := uuid.Parse(formulaID)
		if err == nil {
			_ = fid
			// re-run with formula filter
			query = s.db.Raw(`
				SELECT DISTINCT ON (pps.project_id)
					pps.id          AS score_id,
					pps.project_id,
					p.name          AS project_name,
					p.status        AS project_status,
					pps.formula_id,
					pf.name         AS formula_name,
					pf.version      AS formula_version,
					pps.total_score,
					pps.score_category,
					pps.rank_in_org,
					pps.missing_components,
					pps.calculated_at
				FROM project_priority_scores pps
				JOIN projects p       ON p.id = pps.project_id AND p.deleted_at IS NULL
				JOIN priority_formulas pf ON pf.id = pps.formula_id AND pf.deleted_at IS NULL
				WHERE pps.organization_id = ?
				  AND p.organization_id   = ?
				  AND pps.formula_id      = ?
			`, org, org, fid)
		}
	}

	var rows []ProjectScoreSummary
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	// Sort by total_score DESC
	sortProjectScores(rows)

	counts := map[string]int{
		CategoryLow:      0,
		CategoryMedium:   0,
		CategoryHigh:     0,
		CategoryCritical: 0,
	}
	for _, r := range rows {
		counts[r.ScoreCategory]++
	}

	return &RankingResponse{Counts: counts, Projects: rows}, nil
}

// ------------------------------------------------------------------
// Internal helpers
// ------------------------------------------------------------------

func (s *Service) resolveFormula(org uuid.UUID, formulaID string) (*Formula, error) {
	if formulaID != "" {
		id, err := uuid.Parse(formulaID)
		if err != nil {
			return nil, errors.New("invalid formula_id")
		}
		return s.GetFormula(org, id)
	}
	var f Formula
	err := s.db.Where("organization_id = ? AND status = ? AND deleted_at IS NULL", org, FormulaStatusActive).
		Preload("Components").First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNoActiveFormula
	}
	return &f, err
}

func (s *Service) verifyProject(org, projectID uuid.UUID) error {
	var count int64
	s.db.Table("projects").
		Where("organization_id = ? AND id = ? AND deleted_at IS NULL", org, projectID).
		Count(&count)
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

// computeScore fetches all component values and saves a Score + ScoreComponents.
func (s *Service) computeScore(org, actor, projectID uuid.UUID, f *Formula) (*Score, error) {
	var thresholds map[string]CategoryThreshold
	_ = json.Unmarshal(f.CategoryThresholds, &thresholds)
	if thresholds == nil {
		thresholds = defaultThresholds()
	}

	// Build weight map from formula components
	weightMap := map[string]float64{}
	labelMap := map[string]string{}
	for _, c := range f.Components {
		weightMap[c.ComponentKey] = c.Weight
		labelMap[c.ComponentKey] = c.Label
	}

	now := time.Now()
	sc := &Score{
		ID:             uuid.New(),
		OrganizationID: org,
		ProjectID:      projectID,
		FormulaID:      f.ID,
		FormulaVersion: f.Version,
		CalculatedAt:   now,
		CalculatedBy:   &actor,
		CreatedAt:      now,
	}

	type compResult struct {
		rawValue        *float64
		rawUnit         string
		normalizedScore *float64
		available       bool
		note            string
	}

	componentResults := map[string]compResult{}
	for _, key := range ComponentKeys {
		rv, unit, norm, available, note := s.fetchComponent(org, projectID, key)
		componentResults[key] = compResult{
			rawValue:        rv,
			rawUnit:         unit,
			normalizedScore: norm,
			available:       available,
			note:            note,
		}
	}

	totalWeightedScore := 0.0
	totalWeight := 0.0
	missingCount := 0

	// Persist score first (components reference it)
	if err := s.db.Create(sc).Error; err != nil {
		return nil, err
	}

	for _, key := range ComponentKeys {
		w := weightMap[key]
		if w <= 0 {
			continue // this component not in formula
		}
		cr := componentResults[key]
		label := labelMap[key]
		if label == "" {
			label = key
		}

		comp := &ScoreComponent{
			ID:              uuid.New(),
			ScoreID:         sc.ID,
			OrganizationID:  org,
			ProjectID:       projectID,
			ComponentKey:    key,
			Label:           label,
			RawValue:        cr.rawValue,
			RawUnit:         cr.rawUnit,
			NormalizedScore: cr.normalizedScore,
			Weight:          w,
			Available:       cr.available,
			Note:            cr.note,
			CreatedAt:       now,
		}

		if cr.available && cr.normalizedScore != nil {
			ws := (*cr.normalizedScore) * w
			comp.WeightedScore = &ws
			totalWeightedScore += ws
			totalWeight += w
		} else {
			missingCount++
			switch f.MissingDataRule {
			case MissingPenalize:
				// treat missing as 0 score, but still count weight
				ws := 0.0
				comp.WeightedScore = &ws
				totalWeight += w
			case MissingNeutral:
				// treat as 50 normalized
				ns := 50.0
				ws := ns * w
				comp.NormalizedScore = &ns
				comp.WeightedScore = &ws
				totalWeightedScore += ws
				totalWeight += w
			case MissingExclude:
				// exclude from denominator entirely
			}
		}

		if err := s.db.Create(comp).Error; err != nil {
			return nil, err
		}
	}

	// Calculate total score (0-100)
	finalScore := 0.0
	if totalWeight > 0 {
		finalScore = math.Round((totalWeightedScore/totalWeight)*100*100) / 100
		if f.MissingDataRule == MissingPenalize {
			finalScore = math.Round(totalWeightedScore*100*100) / 100
		}
	}
	finalScore = math.Min(math.Max(finalScore, 0), 100)

	category := classifyScore(finalScore, thresholds)
	sc.TotalScore = finalScore
	sc.ScoreCategory = category
	sc.MissingComponents = missingCount

	if err := s.db.Save(sc).Error; err != nil {
		return nil, err
	}
	return sc, nil
}

// fetchComponent retrieves the raw value for a scoring dimension and normalizes it to 0-100.
// Higher normalized score = higher priority concern (worse situation = higher urgency).
func (s *Service) fetchComponent(org, projectID uuid.UUID, key string) (rawValue *float64, unit string, normalizedScore *float64, available bool, note string) {
	switch key {

	case "health_score":
		// Latest health snapshot score (0-100 inverted: lower health = higher priority)
		type row struct {
			Score float64
		}
		var r row
		err := s.db.Raw(`
			SELECT score FROM health_snapshots
			WHERE organization_id = ? AND project_id = ?
			ORDER BY calculated_at DESC LIMIT 1
		`, org, projectID).Scan(&r).Error
		if err != nil || r.Score == 0 {
			return nil, "", nil, false, "no health snapshot available"
		}
		v := r.Score
		rawValue = &v
		// Invert: health 100 (perfect) → priority 0; health 0 (critical) → priority 100
		norm := math.Round((100-v)*100) / 100
		return rawValue, "score", &norm, true, "from latest health snapshot (inverted: lower health = higher priority)"

	case "risk_score":
		// Max risk_score among open risks (1-25 scale, normalize to 0-100)
		type row struct {
			MaxScore float64
			Count    int
		}
		var r row
		s.db.Raw(`
			SELECT MAX(risk_score) AS max_score, COUNT(*) AS count
			FROM risks
			WHERE organization_id = ? AND project_id = ? AND status NOT IN ('CLOSED','RESOLVED') AND deleted_at IS NULL
		`, org, projectID).Scan(&r)
		if r.Count == 0 {
			return nil, "", nil, false, "no open risks found"
		}
		v := r.MaxScore
		rawValue = &v
		norm := math.Round(math.Min(v/25.0*100, 100)*100) / 100
		return rawValue, "risk_score", &norm, true, fmt.Sprintf("max risk_score among %d open risk(s)", r.Count)

	case "issue_severity":
		// Count of open CRITICAL+HIGH issues, normalized against all open issues
		type row struct {
			Total    int
			Critical int
			High     int
		}
		var r row
		s.db.Raw(`
			SELECT
				COUNT(*) AS total,
				COUNT(*) FILTER (WHERE severity = 'CRITICAL') AS critical,
				COUNT(*) FILTER (WHERE severity = 'HIGH')     AS high
			FROM issues
			WHERE organization_id = ? AND project_id = ?
			  AND status NOT IN ('CLOSED','RESOLVED') AND deleted_at IS NULL
		`, org, projectID).Scan(&r)
		if r.Total == 0 {
			return nil, "", nil, false, "no open issues found"
		}
		v := float64(r.Total)
		rawValue = &v
		// Weight: CRITICAL=1.0, HIGH=0.6; normalize to 0-100 capped at 10 issues max
		weighted := float64(r.Critical)*1.0 + float64(r.High)*0.6
		norm := math.Round(math.Min(weighted/10.0*100, 100)*100) / 100
		return rawValue, "count", &norm, true, fmt.Sprintf("%d open issues (%d critical, %d high)", r.Total, r.Critical, r.High)

	case "budget_usage":
		// Aggregate budget usage % across all budget lines
		type row struct {
			TotalPlanned float64
			TotalActual  float64
		}
		var r row
		s.db.Raw(`
			SELECT COALESCE(SUM(planned_amount), 0) AS total_planned,
			       COALESCE(SUM(actual_amount),  0) AS total_actual
			FROM project_budgets
			WHERE project_id = ? AND deleted_at IS NULL
		`, projectID).Scan(&r)
		if r.TotalPlanned <= 0 {
			return nil, "", nil, false, "no budget lines found or planned amount is zero"
		}
		usagePct := r.TotalActual / r.TotalPlanned * 100
		rawValue = &usagePct
		// Normalize: 0-80% → low priority; 80-100% → rising; >100% → CRITICAL
		norm := 0.0
		switch {
		case usagePct >= 100:
			norm = 100
		case usagePct >= 90:
			norm = 80 + (usagePct-90)/10*20
		case usagePct >= 80:
			norm = 50 + (usagePct-80)/10*30
		default:
			norm = usagePct / 80 * 50
		}
		norm = math.Round(norm*100) / 100
		return rawValue, "%", &norm, true, fmt.Sprintf("budget usage %.1f%% (planned=%.0f actual=%.0f)", usagePct, r.TotalPlanned, r.TotalActual)

	case "schedule_variance":
		// Progress vs time elapsed — negative variance = behind schedule
		type row struct {
			Progress  float64
			StartDate *time.Time
			EndDate   *time.Time
		}
		var r row
		s.db.Raw(`
			SELECT progress_pct AS progress, start_date, end_date
			FROM projects WHERE id = ? AND deleted_at IS NULL
		`, projectID).Scan(&r)
		if r.StartDate == nil || r.EndDate == nil {
			return nil, "", nil, false, "project has no start_date or end_date"
		}
		now := time.Now()
		totalDuration := r.EndDate.Sub(*r.StartDate).Hours()
		elapsed := now.Sub(*r.StartDate).Hours()
		if totalDuration <= 0 {
			return nil, "", nil, false, "project duration is zero or negative"
		}
		timeProgress := math.Min(elapsed/totalDuration*100, 100)
		variance := r.Progress - timeProgress // positive = ahead, negative = behind
		rawValue = &variance
		// Invert and normalize: variance -50 → priority 100; variance +50 → priority 0
		norm := math.Round(math.Max(0, math.Min((-variance+50)/100*100, 100))*100) / 100
		return rawValue, "%", &norm, true, fmt.Sprintf("schedule variance %.1f%% (physical=%.1f%% time_elapsed=%.1f%%)", variance, r.Progress, timeProgress)

	case "corrective_action_overdue":
		// Count of overdue + open corrective actions
		type row struct {
			Total   int
			Overdue int
		}
		var r row
		s.db.Raw(`
			SELECT
				COUNT(*) AS total,
				COUNT(*) FILTER (WHERE due_date < NOW() AND status NOT IN ('CLOSED','COMPLETED')) AS overdue
			FROM corrective_actions
			WHERE organization_id = ? AND project_id = ? AND status NOT IN ('CLOSED','COMPLETED') AND deleted_at IS NULL
		`, org, projectID).Scan(&r)
		if r.Total == 0 {
			return nil, "", nil, false, "no open corrective actions found"
		}
		v := float64(r.Overdue)
		rawValue = &v
		norm := math.Round(math.Min(v/5.0*100, 100)*100) / 100
		return rawValue, "count", &norm, true, fmt.Sprintf("%d overdue out of %d open corrective actions", r.Overdue, r.Total)

	case "benefit_indicator":
		// Aggregate benefit achievement ratio: actual vs target
		type row struct {
			Count  int
			AvgAch float64
		}
		var r row
		s.db.Raw(`
			SELECT
				COUNT(bm.id) AS count,
				COALESCE(AVG(CASE WHEN bm.target > 0 THEN bm.actual / bm.target ELSE NULL END), 0) AS avg_ach
			FROM benefit_measurements bm
			JOIN benefit_indicators bi ON bi.id = bm.indicator_id AND bi.deleted_at IS NULL
			WHERE bi.organization_id = ? AND bi.project_id = ? AND bm.deleted_at IS NULL
		`, org, projectID).Scan(&r)
		if r.Count == 0 {
			return nil, "", nil, false, "no benefit measurements found"
		}
		achPct := r.AvgAch * 100
		rawValue = &achPct
		// Invert: low achievement = high priority concern
		norm := math.Round(math.Max(0, 100-achPct)*100) / 100
		return rawValue, "%", &norm, true, fmt.Sprintf("average benefit achievement %.1f%% from %d measurement(s)", achPct, r.Count)

	default:
		return nil, "", nil, false, fmt.Sprintf("unknown component key: %s", key)
	}
}

func (s *Service) record(org, actor uuid.UUID, action, entityType, entityID string) {
	if s.audit == nil {
		return
	}
	s.audit.Record(audit.WriteRequest{
		OrganizationID: org,
		ActorID:        &actor,
		Action:         action,
		EntityType:     entityType,
		EntityID:       entityID,
	})
}

// ------------------------------------------------------------------
// Pure helpers
// ------------------------------------------------------------------

func validateWeights(comps []ComponentWeightInput) error {
	if len(comps) == 0 {
		return errors.New("at least one component is required")
	}
	sum := 0.0
	for _, c := range comps {
		sum += c.Weight
	}
	if math.Abs(sum-1.0) > 0.01 {
		return fmt.Errorf("%w: got %.4f", ErrWeightSumInvalid, sum)
	}
	return nil
}

func defaultThresholds() map[string]CategoryThreshold {
	return map[string]CategoryThreshold{
		CategoryLow:      {Min: 0, Max: 39},
		CategoryMedium:   {Min: 40, Max: 59},
		CategoryHigh:     {Min: 60, Max: 79},
		CategoryCritical: {Min: 80, Max: 100},
	}
}

func classifyScore(score float64, thresholds map[string]CategoryThreshold) string {
	for _, cat := range []string{CategoryCritical, CategoryHigh, CategoryMedium, CategoryLow} {
		if t, ok := thresholds[cat]; ok {
			if score >= t.Min && score <= t.Max {
				return cat
			}
		}
	}
	if score > 100 {
		return CategoryCritical
	}
	return CategoryLow
}

func sortProjectScores(rows []ProjectScoreSummary) {
	// Simple insertion sort descending by TotalScore
	for i := 1; i < len(rows); i++ {
		for j := i; j > 0 && rows[j].TotalScore > rows[j-1].TotalScore; j-- {
			rows[j], rows[j-1] = rows[j-1], rows[j]
		}
	}
}
