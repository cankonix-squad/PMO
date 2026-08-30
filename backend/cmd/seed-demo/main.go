package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/core/organization"
	projectmodule "github.com/harmanto-49/cankora/internal/modules/project"
	"github.com/harmanto-49/cankora/internal/platform/config"
	"github.com/harmanto-49/cankora/internal/platform/database"
	"github.com/harmanto-49/cankora/internal/shared/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type projectSeed struct {
	Code           string
	Name           string
	Description    string
	Category       string
	Status         string
	Priority       string
	BudgetTotal    float64
	ProgressPct    float64
	StartDaysAgo   int
	EndDaysFromNow int
	IssueTitle     string
	IssueSeverity  string
	RiskTitle      string
	Probability    int
	Impact         int
	BudgetUsage    float64
	OrgUnitCode    string
}

type orgUnitSeed struct {
	Code  string
	Name  string
	Level organization.OrgUnitLevel
}

var demoOrgUnits = []orgUnitSeed{
	{"BBWS-CILIWUNG-CISADANE", "BBWS Ciliwung Cisadane", 3},
	{"BBWS-CITANDUY", "BBWS Citanduy", 3},
	{"BWS-SUMATERA-VII", "BWS Sumatera VII", 3},
	{"BBWS-SERAYU-OPAK", "BBWS Serayu Opak", 3},
	{"BWS-NUSA-TENGGARA-II", "BWS Nusa Tenggara II", 3},
}

var demoProjects = []projectSeed{
	{"DEMO-SDA-001", "Bendungan Marga Tiga", "Pembangunan bendungan untuk ketahanan air dan irigasi Lampung.", "BENDUNGAN", "ACTIVE", "CRITICAL", 2_730_000_000_000, 47.6, 420, 8, "Pembebasan lahan belum tuntas", "CRITICAL", "Cuaca ekstrem menghambat pekerjaan", 5, 4, 1.04, "BWS-SUMATERA-VII"},
	{"DEMO-SDA-002", "SPAM Regional Jatiluhur I", "Pengembangan sistem penyediaan air minum regional.", "AIR_BAKU", "ACTIVE", "CRITICAL", 1_890_000_000_000, 41.2, 360, 11, "Finalisasi skema pendanaan", "HIGH", "Koordinasi lintas instansi terlambat", 4, 4, 0.96, "BBWS-CITANDUY"},
	{"DEMO-SDA-003", "Normalisasi Sungai Ciliwung", "Normalisasi sungai dan pengendalian banjir kawasan perkotaan.", "PENGENDALIAN_BANJIR", "ON_HOLD", "HIGH", 1_240_000_000_000, 50.3, 300, -9, "Permukiman sekitar bantaran", "HIGH", "Akses pekerjaan terbatas", 4, 3, 0.82, "BBWS-CILIWUNG-CISADANE"},
	{"DEMO-SDA-004", "Bendungan Bener", "Pembangunan bendungan multipurpose di Jawa Tengah.", "BENDUNGAN", "ACTIVE", "HIGH", 1_550_000_000_000, 53.1, 500, 13, "Ketersediaan material utama", "HIGH", "Gangguan rantai pasok material", 4, 4, 0.93, "BBWS-SERAYU-OPAK"},
	{"DEMO-SDA-005", "Jaringan Irigasi D.I. Rentang", "Rehabilitasi dan peningkatan layanan jaringan irigasi.", "IRIGASI", "ACTIVE", "HIGH", 912_000_000_000, 60.2, 280, 24, "Perubahan desain saluran primer", "MEDIUM", "Kinerja kontraktor perlu ditingkatkan", 3, 4, 0.74, "BBWS-CITANDUY"},
	{"DEMO-SDA-006", "Bendungan Karian", "Penyelesaian fasilitas pendukung dan operasional bendungan.", "BENDUNGAN", "ACTIVE", "MEDIUM", 2_310_000_000_000, 78.4, 620, 38, "Integrasi sistem operasi bendungan", "MEDIUM", "Kesiapan operasi belum menyeluruh", 3, 3, 0.88, "BBWS-CILIWUNG-CISADANE"},
	{"DEMO-SDA-007", "Pengendalian Banjir Serayu", "Penguatan prasarana pengendalian banjir DAS Serayu.", "PENGENDALIAN_BANJIR", "PLANNING", "HIGH", 685_000_000_000, 18.7, 90, 120, "Dokumen lingkungan perlu dilengkapi", "MEDIUM", "Perizinan berpotensi melewati jadwal", 4, 3, 0.35, "BBWS-SERAYU-OPAK"},
	{"DEMO-SDA-008", "Sistem Air Baku Sepaku", "Pengembangan sistem air baku untuk kawasan layanan baru.", "AIR_BAKU", "ACTIVE", "MEDIUM", 1_120_000_000_000, 69.5, 410, 55, "Sinkronisasi jadwal pengujian", "LOW", "Perubahan kebutuhan kapasitas layanan", 2, 3, 0.67, "BWS-NUSA-TENGGARA-II"},
	{"DEMO-SDA-009", "Rehabilitasi Irigasi Batanghari", "Modernisasi jaringan dan peningkatan efisiensi irigasi.", "IRIGASI", "COMPLETED", "LOW", 540_000_000_000, 100, 700, -30, "Dokumen serah terima akhir", "LOW", "Masa pemeliharaan pascakonstruksi", 2, 2, 1.0, "BWS-SUMATERA-VII"},
	{"DEMO-SDA-010", "Embung Rote Ndao", "Penyediaan tampungan air untuk wilayah rawan kekeringan.", "BENDUNGAN", "DRAFT", "MEDIUM", 325_000_000_000, 5, 20, 240, "Konfirmasi kebutuhan lahan", "MEDIUM", "Ketersediaan lahan belum pasti", 3, 3, 0.1, "BWS-NUSA-TENGGARA-II"},
}

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() //nolint:errcheck

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("seed-demo: config: %v", err)
	}
	db, err := database.Connect(cfg.Database.DSN, false, logger)
	if err != nil {
		log.Fatalf("seed-demo: database: %v", err)
	}
	defer database.Close(db)

	ctx := context.Background()
	var org organization.Organization
	if err := db.WithContext(ctx).Where("code = ?", "CANKORA").First(&org).Error; err != nil {
		log.Fatalf("seed-demo: organization not found; run make seed first: %v", err)
	}
	var admin auth.User
	if err := db.WithContext(ctx).Where("email = ?", "admin@cankora.local").First(&admin).Error; err != nil {
		log.Fatalf("seed-demo: admin not found; run make seed first: %v", err)
	}

	// Seed demo org units (BBWS/BWS) idempotent
	orgUnitMap := make(map[string]uuid.UUID) // code → id
	for _, ou := range demoOrgUnits {
		id, err := upsertOrgUnit(db, ctx, org.ID, ou)
		if err != nil {
			log.Fatalf("seed-demo: org unit %s: %v", ou.Code, err)
		}
		orgUnitMap[ou.Code] = id
	}

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for index, seed := range demoProjects {
			var orgUnitID *uuid.UUID
			if seed.OrgUnitCode != "" {
				if id, ok := orgUnitMap[seed.OrgUnitCode]; ok {
					orgUnitID = &id
				}
			}
			project, err := upsertProject(tx, org.ID, admin.ID, seed, orgUnitID)
			if err != nil {
				return err
			}
			if err := upsertChildren(tx, org.ID, admin.ID, project, seed, index); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Fatalf("seed-demo: transaction: %v", err)
	}

	logger.Info("demo data seeded", zap.Int("projects", len(demoProjects)))
}

func upsertOrgUnit(db *gorm.DB, ctx context.Context, orgID uuid.UUID, seed orgUnitSeed) (uuid.UUID, error) {
	var unit organization.OrgUnit
	err := db.WithContext(ctx).Unscoped().
		Where("organization_id = ? AND code = ?", orgID, seed.Code).
		First(&unit).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		unit = organization.OrgUnit{
			ID:             uuid.New(),
			OrganizationID: orgID,
			Code:           seed.Code,
			Name:           seed.Name,
			Level:          seed.Level,
			IsActive:       true,
		}
		if err := db.WithContext(ctx).Create(&unit).Error; err != nil {
			return uuid.Nil, err
		}
		return unit.ID, nil
	}
	// Update name/level in case it changed
	unit.Name = seed.Name
	unit.Level = seed.Level
	unit.IsActive = true
	unit.DeletedAt = gorm.DeletedAt{}
	return unit.ID, db.WithContext(ctx).Save(&unit).Error
}

func upsertProject(db *gorm.DB, orgID, adminID uuid.UUID, seed projectSeed, orgUnitID *uuid.UUID) (*projectmodule.Project, error) {
	var item projectmodule.Project
	err := db.Unscoped().Where("organization_id = ? AND code = ?", orgID, seed.Code).First(&item).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item = projectmodule.Project{
			ID:             uuid.New(),
			OrganizationID: orgID,
			Code:           seed.Code,
			CreatedBy:      adminID,
		}
	}
	item.Name = seed.Name
	item.Description = seed.Description
	item.Objectives = "Meningkatkan layanan sumber daya air yang efektif, tepat waktu, dan akuntabel."
	item.Category = seed.Category
	item.Status = seed.Status
	item.Priority = seed.Priority
	item.BudgetTotal = seed.BudgetTotal
	item.Currency = "IDR"
	item.ProgressPct = seed.ProgressPct
	item.ManagerID = &adminID
	item.OrgUnitID = orgUnitID
	item.StartDate = flexDate(-seed.StartDaysAgo)
	item.EndDate = flexDate(seed.EndDaysFromNow)
	item.DeletedAt = gorm.DeletedAt{}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &item, db.Create(&item).Error
	}
	return &item, db.Save(&item).Error
}

func upsertChildren(db *gorm.DB, orgID, adminID uuid.UUID, project *projectmodule.Project, seed projectSeed, index int) error {
	milestone := projectmodule.Milestone{
		OrganizationID: orgID,
		ProjectID:      project.ID,
		Title:          "Target pekerjaan utama",
		Description:    "Milestone demo untuk monitoring target proyek.",
		DueDate:        flexDate(seed.EndDaysFromNow - 5),
		Status:         "IN_PROGRESS",
		ProgressPct:    seed.ProgressPct,
		CreatedBy:      adminID,
	}
	if index < 4 {
		milestone.DueDate = flexDate(-3 - index)
		milestone.Status = "DELAYED"
	}
	if err := upsertMilestone(db, &milestone); err != nil {
		return err
	}

	taskStatuses := []string{"IN_PROGRESS", "TODO"}
	for taskIndex, status := range taskStatuses {
		task := projectmodule.Task{
			OrganizationID: orgID,
			ProjectID:      project.ID,
			MilestoneID:    &milestone.ID,
			WBSCode:        "DEMO-" + seed.Code + "-" + string(rune('A'+taskIndex)),
			Title:          []string{"Penyelesaian pekerjaan konstruksi utama", "Verifikasi dokumen dan bukti lapangan"}[taskIndex],
			Description:    "Task demo CANKORA untuk monitoring delivery.",
			Status:         status,
			Priority:       seed.Priority,
			Type:           "TASK",
			StartDate:      flexDate(-45 + taskIndex*10),
			DueDate:        flexDate(20 + taskIndex*12),
			EstHours:       160,
			ActualHours:    72,
			ProgressPct:    seed.ProgressPct,
			CreatedBy:      adminID,
		}
		if taskIndex == 0 && index < 5 {
			task.DueDate = flexDate(-2 - index)
		}
		if err := upsertTask(db, &task); err != nil {
			return err
		}
	}

	issue := projectmodule.Issue{
		OrganizationID: orgID,
		ProjectID:      project.ID,
		Title:          seed.IssueTitle,
		Description:    "Isu demo untuk kebutuhan dashboard eksekutif.",
		Status:         "OPEN",
		Severity:       seed.IssueSeverity,
		Escalation:     "EXECUTIVE",
		ReportedBy:     adminID,
		AssignedTo:     &adminID,
		DueDate:        flexDate(14 + index),
	}
	if err := upsertIssue(db, &issue); err != nil {
		return err
	}

	risk := projectmodule.Risk{
		OrganizationID: orgID,
		ProjectID:      project.ID,
		Title:          seed.RiskTitle,
		Description:    "Risiko demo untuk register risiko proyek.",
		Status:         "ASSESSED",
		Probability:    seed.Probability,
		Impact:         seed.Impact,
		RiskScore:      projectmodule.RiskScore(seed.Probability, seed.Impact),
		Severity:       projectmodule.RiskSeverity(seed.Probability, seed.Impact),
		Mitigation:     "Percepat koordinasi, lakukan monitoring mingguan, dan eskalasi bila target terlewati.",
		OwnedBy:        &adminID,
		DueDate:        flexDate(30 + index),
		CreatedBy:      adminID,
	}
	if err := upsertRisk(db, &risk); err != nil {
		return err
	}

	budget := projectmodule.ProjectBudget{
		ProjectID:   project.ID,
		Category:    "Konstruksi Utama",
		Description: "Anggaran demo untuk monitoring penyerapan.",
		Planned:     seed.BudgetTotal,
		Actual:      seed.BudgetTotal * seed.BudgetUsage,
		Currency:    "IDR",
		CreatedBy:   adminID,
	}
	if err := upsertBudget(db, &budget); err != nil {
		return err
	}

	for month := 4; month >= 0; month-- {
		progress := projectmodule.ProgressHistory{
			ProjectID:   project.ID,
			ProgressPct: maxFloat(0, seed.ProgressPct-float64(month*7)),
			Notes:       "Snapshot progres demo bulanan.",
			RecordedBy:  adminID,
			RecordedAt:  time.Now().AddDate(0, -month, 0),
		}
		if err := upsertProgress(db, &progress); err != nil {
			return err
		}
	}

	// Seed 6 months of periodic reports for dashboard trend
	now := time.Now()
	for monthOffset := 5; monthOffset >= 0; monthOffset-- {
		t := now.AddDate(0, -monthOffset, 0)
		physical := maxFloat(0, seed.ProgressPct-float64(monthOffset*8))
		planned := seed.BudgetTotal / 6
		actual := planned * maxFloat(0.1, seed.BudgetUsage-float64(monthOffset)*0.05)
		report := projectmodule.PeriodicReport{
			OrganizationID:      orgID,
			ProjectID:           project.ID,
			PeriodYear:          t.Year(),
			PeriodMonth:         int(t.Month()),
			PhysicalProgressPct: physical,
			FinancialPlanned:    planned,
			FinancialActual:     actual,
			FinancialPct:        actual / planned * 100,
			Notes:               "Laporan periodik demo bulanan.",
			ReportedBy:          &adminID,
			ReportedAt:          t,
		}
		if err := upsertPeriodicReport(db, orgID, &report); err != nil {
			return err
		}
	}

	return nil
}

func upsertMilestone(db *gorm.DB, item *projectmodule.Milestone) error {
	var existing projectmodule.Milestone
	err := db.Unscoped().Where("project_id = ? AND title = ?", item.ProjectID, item.Title).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item.ID = uuid.New()
		return db.Create(item).Error
	}
	if err != nil {
		return err
	}
	item.ID = existing.ID
	item.CreatedAt = existing.CreatedAt
	item.DeletedAt = gorm.DeletedAt{}
	return db.Save(item).Error
}

func upsertTask(db *gorm.DB, item *projectmodule.Task) error {
	var existing projectmodule.Task
	err := db.Unscoped().Where("project_id = ? AND wbs_code = ?", item.ProjectID, item.WBSCode).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item.ID = uuid.New()
		return db.Create(item).Error
	}
	if err != nil {
		return err
	}
	item.ID = existing.ID
	item.CreatedAt = existing.CreatedAt
	item.DeletedAt = gorm.DeletedAt{}
	return db.Save(item).Error
}

func upsertIssue(db *gorm.DB, item *projectmodule.Issue) error {
	var existing projectmodule.Issue
	err := db.Unscoped().Where("project_id = ? AND title = ?", item.ProjectID, item.Title).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item.ID = uuid.New()
		return db.Create(item).Error
	}
	if err != nil {
		return err
	}
	item.ID = existing.ID
	item.CreatedAt = existing.CreatedAt
	item.DeletedAt = gorm.DeletedAt{}
	return db.Save(item).Error
}

func upsertRisk(db *gorm.DB, item *projectmodule.Risk) error {
	var existing projectmodule.Risk
	err := db.Unscoped().Where("project_id = ? AND title = ?", item.ProjectID, item.Title).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item.ID = uuid.New()
		return db.Create(item).Error
	}
	if err != nil {
		return err
	}
	item.ID = existing.ID
	item.CreatedAt = existing.CreatedAt
	item.DeletedAt = gorm.DeletedAt{}
	return db.Save(item).Error
}

func upsertBudget(db *gorm.DB, item *projectmodule.ProjectBudget) error {
	var existing projectmodule.ProjectBudget
	err := db.Unscoped().Where("project_id = ? AND category = ?", item.ProjectID, item.Category).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item.ID = uuid.New()
		return db.Create(item).Error
	}
	if err != nil {
		return err
	}
	item.ID = existing.ID
	item.CreatedAt = existing.CreatedAt
	item.DeletedAt = gorm.DeletedAt{}
	return db.Save(item).Error
}

func upsertProgress(db *gorm.DB, item *projectmodule.ProgressHistory) error {
	monthStart := time.Date(item.RecordedAt.Year(), item.RecordedAt.Month(), 1, 0, 0, 0, 0, item.RecordedAt.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)
	var existing projectmodule.ProgressHistory
	err := db.Where("project_id = ? AND recorded_at >= ? AND recorded_at < ? AND notes = ?", item.ProjectID, monthStart, monthEnd, item.Notes).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item.ID = uuid.New()
		return db.Create(item).Error
	}
	if err != nil {
		return err
	}
	item.ID = existing.ID
	return db.Save(item).Error
}

func upsertPeriodicReport(db *gorm.DB, orgID uuid.UUID, item *projectmodule.PeriodicReport) error {
	var existing projectmodule.PeriodicReport
	err := db.Unscoped().
		Where("organization_id = ? AND project_id = ? AND period_year = ? AND period_month = ?",
			orgID, item.ProjectID, item.PeriodYear, item.PeriodMonth).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item.ID = uuid.New()
		return db.Create(item).Error
	}
	if err != nil {
		return err
	}
	item.ID = existing.ID
	item.CreatedAt = existing.CreatedAt
	item.DeletedAt = gorm.DeletedAt{}
	return db.Save(item).Error
}

func flexDate(daysFromNow int) *types.FlexTime {
	return &types.FlexTime{Time: time.Now().AddDate(0, 0, daysFromNow)}
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
