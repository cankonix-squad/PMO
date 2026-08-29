package imports

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ParsedRow holds the raw header→value map for one CSV row.
type ParsedRow struct {
	RowNumber int
	Data      map[string]string
}

// ParseCSV reads all data rows from a CSV reader.
// It expects the first row to be a header. Returns ParsedRow slice.
func ParseCSV(r io.Reader) ([]ParsedRow, error) {
	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true
	cr.FieldsPerRecord = -1 // allow variable fields

	headers, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	// Normalize headers: trim + lowercase
	for i, h := range headers {
		headers[i] = strings.ToLower(strings.TrimSpace(h))
	}
	if len(headers) == 0 {
		return nil, fmt.Errorf("CSV has no header row")
	}

	var rows []ParsedRow
	rowNum := 1 // 1-based, header was row 0
	for {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row %d: %w", rowNum, err)
		}
		rowNum++

		// Skip completely blank rows
		allBlank := true
		for _, v := range record {
			if strings.TrimSpace(v) != "" {
				allBlank = false
				break
			}
		}
		if allBlank {
			continue
		}

		data := make(map[string]string, len(headers))
		for i, h := range headers {
			if i < len(record) {
				data[h] = strings.TrimSpace(record[i])
			} else {
				data[h] = ""
			}
		}
		rows = append(rows, ParsedRow{RowNumber: rowNum - 1, Data: data})
	}
	return rows, nil
}

// marshalJSON is a helper to JSON-encode a value to string for DB storage.
func marshalJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// marshalErrors encodes a string slice as JSON array.
func marshalErrors(errs []string) string {
	b, err := json.Marshal(errs)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// Templates returns all supported dataset template definitions.
func Templates() []TemplateDef {
	return []TemplateDef{
		{
			DatasetType: DatasetProjectProgress,
			DisplayName: "Project Progress Update",
			Description: "Update physical progress percentage per project per period.",
			Columns: []ColumnDef{
				{Name: "project_code", Required: true, Type: "string", Description: "Kode unik proyek", Example: "PRJ-001"},
				{Name: "progress_pct", Required: true, Type: "number", Description: "Persentase kemajuan fisik (0-100)", Example: "75.5"},
				{Name: "period_date", Required: false, Type: "date", Description: "Tanggal periode (YYYY-MM-DD)", Example: "2026-08-01"},
				{Name: "note", Required: false, Type: "string", Description: "Catatan progres", Example: "On track"},
			},
		},
		{
			DatasetType: DatasetProjectBudgets,
			DisplayName: "Project Budget Lines",
			Description: "Import or update budget category lines per project.",
			Columns: []ColumnDef{
				{Name: "project_code", Required: true, Type: "string", Description: "Kode unik proyek", Example: "PRJ-001"},
				{Name: "category", Required: true, Type: "string", Description: "Kategori anggaran", Example: "CONSTRUCTION"},
				{Name: "planned", Required: false, Type: "number", Description: "Anggaran rencana (IDR)", Example: "5000000000"},
				{Name: "actual", Required: false, Type: "number", Description: "Realisasi anggaran (IDR)", Example: "3200000000"},
				{Name: "currency", Required: false, Type: "string", Description: "Mata uang (default IDR)", Example: "IDR"},
			},
		},
		{
			DatasetType: DatasetRisks,
			DisplayName: "Project Risks",
			Description: "Import risks per project. Backend computes risk_score and severity.",
			Columns: []ColumnDef{
				{Name: "project_code", Required: true, Type: "string", Description: "Kode unik proyek", Example: "PRJ-001"},
				{Name: "title", Required: true, Type: "string", Description: "Judul risiko", Example: "Keterlambatan pengadaan"},
				{Name: "description", Required: false, Type: "string", Description: "Deskripsi risiko"},
				{Name: "probability", Required: true, Type: "number", Description: "Probabilitas 1-5", Example: "3"},
				{Name: "impact", Required: true, Type: "number", Description: "Dampak 1-5", Example: "4"},
				{Name: "mitigation", Required: false, Type: "string", Description: "Rencana mitigasi"},
				{Name: "owner", Required: false, Type: "string", Description: "PIC/owner risiko"},
				{Name: "due_date", Required: false, Type: "date", Description: "Tanggal target mitigasi (YYYY-MM-DD)"},
			},
		},
		{
			DatasetType: DatasetIssues,
			DisplayName: "Project Issues",
			Description: "Import issues per project.",
			Columns: []ColumnDef{
				{Name: "project_code", Required: true, Type: "string", Description: "Kode unik proyek", Example: "PRJ-001"},
				{Name: "title", Required: true, Type: "string", Description: "Judul isu", Example: "Lahan belum bebas"},
				{Name: "description", Required: false, Type: "string", Description: "Deskripsi isu"},
				{Name: "severity", Required: false, Type: "enum", Description: "CRITICAL|HIGH|MEDIUM|LOW", Example: "HIGH"},
				{Name: "due_date", Required: false, Type: "date", Description: "Tanggal target penyelesaian (YYYY-MM-DD)"},
				{Name: "owner", Required: false, Type: "string", Description: "PIC/owner isu"},
			},
		},
		{
			DatasetType: DatasetBenefitMeasurements,
			DisplayName: "Benefit Measurements",
			Description: "Import actual benefit measurements per indicator per period.",
			Columns: []ColumnDef{
				{Name: "project_code", Required: true, Type: "string", Description: "Kode unik proyek", Example: "PRJ-001"},
				{Name: "indicator_name", Required: true, Type: "string", Description: "Nama indikator manfaat", Example: "Luas irigasi (ha)"},
				{Name: "period_year", Required: true, Type: "number", Description: "Tahun periode", Example: "2026"},
				{Name: "period_month", Required: false, Type: "number", Description: "Bulan periode 1-12 (default 1)", Example: "8"},
				{Name: "actual_value", Required: true, Type: "number", Description: "Nilai aktual", Example: "1250.5"},
				{Name: "target_value", Required: false, Type: "number", Description: "Nilai target"},
			},
		},
	}
}

// TemplateByType returns the TemplateDef for a given dataset type, or false if not found.
func TemplateByType(datasetType string) (TemplateDef, bool) {
	for _, t := range Templates() {
		if t.DatasetType == datasetType {
			return t, true
		}
	}
	return TemplateDef{}, false
}
