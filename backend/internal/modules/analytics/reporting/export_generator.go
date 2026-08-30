package reporting

// export_generator.go — UAT-002 Report Export Real File
// Generates CSV and XLSX files from reporting read-model datasets.
// File storage: local disk under REPORT_STORAGE_PATH (default: backend/storage/reports).
// storage_key is always relative to REPORT_STORAGE_PATH — never an absolute path.

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// ---------------------------------------------------------------------------
// Storage helpers
// ---------------------------------------------------------------------------

// reportStoragePath returns the absolute directory for report files.
// Uses env REPORT_STORAGE_PATH, falls back to <cwd>/storage/reports.
func reportStoragePath() string {
	if p := os.Getenv("REPORT_STORAGE_PATH"); p != "" {
		return filepath.Clean(p)
	}
	// Resolve relative to working directory (backend/)
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "storage", "reports")
}

// safeStorageKey builds a relative storage key that is path-traversal safe.
// Pattern: <orgID>/<date>/<filename>
func safeStorageKey(orgID uuid.UUID, filename string) string {
	date := time.Now().UTC().Format("2006-01-02")
	// Sanitize filename — only allow alphanum, dash, underscore, dot
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			return r
		}
		return '_'
	}, filename)
	return filepath.Join(orgID.String(), date, safe)
}

// resolveAbsPath converts a relative storage_key to an absolute path,
// rejecting any key that escapes the storage root (path traversal guard).
func resolveAbsPath(storageRoot, storageKey string) (string, error) {
	abs := filepath.Join(storageRoot, filepath.FromSlash(storageKey))
	abs = filepath.Clean(abs)
	root := filepath.Clean(storageRoot)
	if !strings.HasPrefix(abs, root+string(os.PathSeparator)) && abs != root {
		return "", fmt.Errorf("storage key escapes storage root: %s", storageKey)
	}
	return abs, nil
}

// ---------------------------------------------------------------------------
// Generated file result
// ---------------------------------------------------------------------------

type GeneratedFile struct {
	StorageKey  string // relative key stored in DB
	AbsPath     string // full path on disk
	FileName    string // human-readable filename
	MimeType    string
	FileSizeB   int64
	GeneratedAt time.Time
}

// ---------------------------------------------------------------------------
// CSV generation
// ---------------------------------------------------------------------------

func generateCSV(storageRoot string, orgID uuid.UUID, datasetKey string, rows [][]string) (*GeneratedFile, error) {
	filename := fmt.Sprintf("%s_%s.csv", datasetKey, time.Now().UTC().Format("20060102_150405"))
	key := safeStorageKey(orgID, filename)
	absPath, err := resolveAbsPath(storageRoot, key)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o750); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}
	f, err := os.Create(absPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.WriteAll(rows); err != nil {
		return nil, fmt.Errorf("write csv: %w", err)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flush csv: %w", err)
	}

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return &GeneratedFile{
		StorageKey:  filepath.ToSlash(key),
		AbsPath:     absPath,
		FileName:    filename,
		MimeType:    "text/csv; charset=utf-8",
		FileSizeB:   info.Size(),
		GeneratedAt: time.Now().UTC(),
	}, nil
}

// ---------------------------------------------------------------------------
// XLSX generation
// ---------------------------------------------------------------------------

func generateXLSX(storageRoot string, orgID uuid.UUID, datasetKey string, rows [][]string) (*GeneratedFile, error) {
	filename := fmt.Sprintf("%s_%s.xlsx", datasetKey, time.Now().UTC().Format("20060102_150405"))
	key := safeStorageKey(orgID, filename)
	absPath, err := resolveAbsPath(storageRoot, key)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o750); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}

	xl := excelize.NewFile()
	defer xl.Close()
	sheet := "Data"
	xl.SetSheetName("Sheet1", sheet)

	// Header style — bold
	headerStyle, _ := xl.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D9E1F2"}, Pattern: 1},
	})

	for i, row := range rows {
		for j, cell := range row {
			coord, _ := excelize.CoordinatesToCellName(j+1, i+1)
			_ = xl.SetCellValue(sheet, coord, cell)
			if i == 0 {
				_ = xl.SetCellStyle(sheet, coord, coord, headerStyle)
			}
		}
	}
	// Auto-width on first 20 columns
	for col := 1; col <= 20 && col <= len(rows[0]); col++ {
		colName, _ := excelize.ColumnNumberToName(col)
		_ = xl.SetColWidth(sheet, colName, colName, 20)
	}

	if err := xl.SaveAs(absPath); err != nil {
		return nil, fmt.Errorf("save xlsx: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}
	return &GeneratedFile{
		StorageKey:  filepath.ToSlash(key),
		AbsPath:     absPath,
		FileName:    filename,
		MimeType:    "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		FileSizeB:   info.Size(),
		GeneratedAt: time.Now().UTC(),
	}, nil
}

// ---------------------------------------------------------------------------
// Dataset → rows converters (header + data rows)
// ---------------------------------------------------------------------------

func executiveSummaryRows(data *ExecutiveSummaryRow) [][]string {
	return [][]string{
		{"metric", "value"},
		{"total_projects", fmt.Sprintf("%d", data.TotalProjects)},
		{"active_projects", fmt.Sprintf("%d", data.ActiveProjects)},
		{"completed_projects", fmt.Sprintf("%d", data.CompletedProjects)},
		{"on_hold_projects", fmt.Sprintf("%d", data.OnHoldProjects)},
		{"avg_progress_pct", fmt.Sprintf("%.2f", data.AvgProgressPct)},
		{"total_budget_plan", fmt.Sprintf("%.2f", data.TotalBudgetPlan)},
		{"total_budget_actual", fmt.Sprintf("%.2f", data.TotalBudgetActual)},
		{"budget_usage_pct", fmt.Sprintf("%.2f", data.BudgetUsagePct)},
		{"total_risks", fmt.Sprintf("%d", data.TotalRisks)},
		{"open_risks", fmt.Sprintf("%d", data.OpenRisks)},
		{"high_risks", fmt.Sprintf("%d", data.HighRisks)},
		{"total_issues", fmt.Sprintf("%d", data.TotalIssues)},
		{"open_issues", fmt.Sprintf("%d", data.OpenIssues)},
		{"green_health", fmt.Sprintf("%d", data.GreenHealth)},
		{"yellow_health", fmt.Sprintf("%d", data.YellowHealth)},
		{"red_health", fmt.Sprintf("%d", data.RedHealth)},
		{"critical_health", fmt.Sprintf("%d", data.CriticalHealth)},
	}
}

func projectPerformanceRows(data []ProjectPerformanceRow) [][]string {
	rows := [][]string{{
		"project_id", "project_code", "project_name", "status",
		"progress_pct", "start_date", "end_date",
		"budget_plan", "budget_actual", "budget_usage_pct",
		"health_class", "province", "priority_score", "priority_category",
	}}
	for _, r := range data {
		startDate, endDate := "", ""
		if r.StartDate != nil {
			startDate = r.StartDate.Format("2006-01-02")
		}
		if r.EndDate != nil {
			endDate = r.EndDate.Format("2006-01-02")
		}
		healthClass := ""
		if r.HealthClass != nil {
			healthClass = *r.HealthClass
		}
		province := ""
		if r.Province != nil {
			province = *r.Province
		}
		priorityScore := ""
		if r.PriorityScore != nil {
			priorityScore = fmt.Sprintf("%.2f", *r.PriorityScore)
		}
		priorityCategory := ""
		if r.PriorityCategory != nil {
			priorityCategory = *r.PriorityCategory
		}
		rows = append(rows, []string{
			r.ProjectID.String(), r.ProjectCode, r.ProjectName, r.Status,
			fmt.Sprintf("%.2f", r.ProgressPct), startDate, endDate,
			fmt.Sprintf("%.2f", r.BudgetPlan), fmt.Sprintf("%.2f", r.BudgetActual),
			fmt.Sprintf("%.2f", r.BudgetUsagePct),
			healthClass, province, priorityScore, priorityCategory,
		})
	}
	return rows
}

func riskIssueRows(data []RiskIssueRow) [][]string {
	rows := [][]string{{
		"project_id", "project_code", "project_name",
		"total_risks", "open_risks", "high_risks", "critical_risks",
		"total_issues", "open_issues", "high_issues", "critical_issues",
	}}
	for _, r := range data {
		rows = append(rows, []string{
			r.ProjectID.String(), r.ProjectCode, r.ProjectName,
			fmt.Sprintf("%d", r.TotalRisks), fmt.Sprintf("%d", r.OpenRisks),
			fmt.Sprintf("%d", r.HighRisks), fmt.Sprintf("%d", r.CriticalRisks),
			fmt.Sprintf("%d", r.TotalIssues), fmt.Sprintf("%d", r.OpenIssues),
			fmt.Sprintf("%d", r.HighIssues), fmt.Sprintf("%d", r.CriticalIssues),
		})
	}
	return rows
}

func budgetRows(data []BudgetRow) [][]string {
	rows := [][]string{{
		"project_id", "project_code", "project_name", "status",
		"budget_plan", "budget_actual", "variance", "usage_pct",
	}}
	for _, r := range data {
		rows = append(rows, []string{
			r.ProjectID.String(), r.ProjectCode, r.ProjectName, r.Status,
			fmt.Sprintf("%.2f", r.BudgetPlan), fmt.Sprintf("%.2f", r.BudgetActual),
			fmt.Sprintf("%.2f", r.Variance), fmt.Sprintf("%.2f", r.UsagePct),
		})
	}
	return rows
}

func benefitRows(data []BenefitRow) [][]string {
	rows := [][]string{{
		"project_id", "project_code", "project_name",
		"indicator_id", "indicator_name", "unit",
		"target", "actual", "achievement_pct", "aggregation_method",
	}}
	for _, r := range data {
		rows = append(rows, []string{
			r.ProjectID.String(), r.ProjectCode, r.ProjectName,
			r.IndicatorID.String(), r.IndicatorName, r.Unit,
			fmt.Sprintf("%.2f", r.Target), fmt.Sprintf("%.2f", r.Actual),
			fmt.Sprintf("%.2f", r.AchievementPct), r.AggregationMethod,
		})
	}
	return rows
}

func priorityRows(data []PriorityRow) [][]string {
	rows := [][]string{{"project_id", "project_code", "project_name", "total_score", "score_category", "calculated_at"}}
	for _, r := range data {
		rows = append(rows, []string{
			r.ProjectID.String(), r.ProjectCode, r.ProjectName,
			fmt.Sprintf("%.2f", r.TotalScore), r.Category,
			r.CalculatedAt.Format(time.RFC3339),
		})
	}
	return rows
}
