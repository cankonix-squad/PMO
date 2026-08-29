package primavera

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Parsed activity — intermediate representation before DB mapping
// ---------------------------------------------------------------------------

// ParsedActivity is a normalised Primavera P6 activity record, regardless
// of whether the source was XER or PMXML format.
type ParsedActivity struct {
	ActivityID   string `json:"activity_id"`   // P6 act_id / ObjectId
	WBSCode      string `json:"wbs_code"`      // WBS short name
	ActivityName string `json:"activity_name"` // activity name
	PlannedStart string `json:"planned_start"` // ISO date string, may be empty
	PlannedEnd   string `json:"planned_end"`
	ActualStart  string `json:"actual_start"`
	ActualEnd    string `json:"actual_end"`
	// Physical progress: 0–100 scale
	BaselinePhysical float64 `json:"baseline_physical"` // phys_complete_pct at baseline
	ActualPhysical   float64 `json:"actual_physical"`   // current phys_complete_pct
	// Raw source row (JSON-serialised) for lineage
	RawPayload map[string]string `json:"raw_payload"`
}

// ParseResult is returned by all format parsers.
type ParseResult struct {
	Activities []ParsedActivity
	Errors     []SyncErrorEntry
}

// ---------------------------------------------------------------------------
// Dispatcher
// ---------------------------------------------------------------------------

// Parse detects the format and delegates to the appropriate parser.
// format should be FormatXER, FormatPMXML, or FormatJSON.
func Parse(r io.Reader, format string) (*ParseResult, error) {
	switch strings.ToUpper(format) {
	case FormatXER:
		return ParseXER(r)
	case FormatPMXML:
		return ParsePMXML(r)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// ---------------------------------------------------------------------------
// XER parser
// ---------------------------------------------------------------------------
// XER is a tab-delimited text export from Primavera P6.
// Structure:
//   %T <TableName>
//   %F <col1>\t<col2>\t...
//   %R <val1>\t<val2>\t...
//   ...
//   %E   (end of table)

// ParseXER parses a Primavera P6 XER export file.
func ParseXER(r io.Reader) (*ParseResult, error) {
	result := &ParseResult{}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1 MB line buffer

	var currentTable string
	var headers []string
	rowIndex := 0

	// We need two tables: TASK (activities) and PROJWBS (WBS codes).
	// Build lookup maps first, then build ParsedActivity.
	taskRows := []map[string]string{}
	wbsRows := []map[string]string{}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		prefix := ""
		if len(line) >= 2 {
			prefix = line[:2]
		}

		switch prefix {
		case "%T":
			currentTable = strings.TrimSpace(line[2:])
			headers = nil
			rowIndex = 0
		case "%F":
			raw := strings.TrimSpace(line[2:])
			headers = strings.Split(raw, "\t")
		case "%R":
			if len(headers) == 0 {
				continue
			}
			raw := line[2:]
			vals := strings.Split(raw, "\t")
			row := make(map[string]string, len(headers))
			for i, h := range headers {
				if i < len(vals) {
					row[h] = strings.TrimSpace(vals[i])
				} else {
					row[h] = ""
				}
			}
			switch currentTable {
			case "TASK":
				taskRows = append(taskRows, row)
			case "PROJWBS":
				wbsRows = append(wbsRows, row)
			}
			rowIndex++
		case "%E":
			currentTable = ""
			headers = nil
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("xer scan: %w", err)
	}

	// Build WBS lookup: wbs_id → wbs_short_name
	wbsByID := make(map[string]string, len(wbsRows))
	for _, w := range wbsRows {
		if id := w["wbs_id"]; id != "" {
			wbsByID[id] = w["wbs_short_name"]
		}
	}

	// Build activities
	for i, row := range taskRows {
		act := parseXERTaskRow(row, wbsByID, i+1, &result.Errors)
		if act != nil {
			result.Activities = append(result.Activities, *act)
		}
	}
	return result, nil
}

func parseXERTaskRow(row map[string]string, wbsByID map[string]string, rowNum int, errs *[]SyncErrorEntry) *ParsedActivity {
	actID := row["task_code"]
	if actID == "" {
		actID = row["act_id"]
	}
	if actID == "" {
		*errs = append(*errs, SyncErrorEntry{
			Code:    "E001",
			Message: "missing activity id (task_code / act_id)",
			Row:     rowNum,
		})
		return nil
	}

	wbsCode := wbsByID[row["wbs_id"]]

	baselinePhys := parseFloat(row["phys_complete_pct"])
	actualPhys := parseFloat(row["act_phys_complete_pct"])
	if actualPhys == 0 {
		actualPhys = parseFloat(row["phys_complete_pct"])
	}

	return &ParsedActivity{
		ActivityID:       actID,
		WBSCode:          wbsCode,
		ActivityName:     row["task_name"],
		PlannedStart:     normaliseXERDate(row["target_start_date"]),
		PlannedEnd:       normaliseXERDate(row["target_end_date"]),
		ActualStart:      normaliseXERDate(row["act_start_date"]),
		ActualEnd:        normaliseXERDate(row["act_end_date"]),
		BaselinePhysical: clamp100(baselinePhys),
		ActualPhysical:   clamp100(actualPhys),
		RawPayload:       row,
	}
}

// normaliseXERDate converts XER date strings ("2024-01-15 00:00" or "2024-01-15") → "2024-01-15".
func normaliseXERDate(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// try parsing with time component first
	for _, layout := range []string{"2006-01-02 15:04", "2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return s // return as-is; validation will catch it downstream
}

// ---------------------------------------------------------------------------
// PMXML parser
// ---------------------------------------------------------------------------
// PMXML is the Oracle Primavera P6 XML export format.

type pmxmlProject struct {
	XMLName    xml.Name        `xml:"Project"`
	Activities []pmxmlActivity `xml:"Activity"`
	WBSNodes   []pmxmlWBS      `xml:"WBS"`
}

type pmxmlActivity struct {
	ObjectId                      string  `xml:"ObjectId"`
	Id                            string  `xml:"Id"`
	Name                          string  `xml:"Name"`
	WBSObjectId                   string  `xml:"WBSObjectId"`
	PlannedStartDate              string  `xml:"PlannedStartDate"`
	PlannedFinishDate             string  `xml:"PlannedFinishDate"`
	ActualStartDate               string  `xml:"ActualStartDate"`
	ActualFinishDate              string  `xml:"ActualFinishDate"`
	PhysicalPercentComplete       float64 `xml:"PhysicalPercentComplete"`
	ActualPhysicalPercentComplete float64 `xml:"ActualPhysicalPercentComplete"`
}

type pmxmlWBS struct {
	ObjectId string `xml:"ObjectId"`
	Code     string `xml:"Code"`
	Name     string `xml:"Name"`
}

// ParsePMXML parses a Primavera P6 PMXML export.
func ParsePMXML(r io.Reader) (*ParseResult, error) {
	result := &ParseResult{}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("pmxml read: %w", err)
	}

	// The PMXML root is <APIBusinessObjects> containing one or more <Project>
	type pmxmlRoot struct {
		XMLName  xml.Name       `xml:"APIBusinessObjects"`
		Projects []pmxmlProject `xml:"Project"`
	}
	var root pmxmlRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		// Try direct <Project> root
		var single pmxmlProject
		if err2 := xml.Unmarshal(data, &single); err2 != nil {
			return nil, fmt.Errorf("pmxml parse: %w (also tried single-project: %v)", err, err2)
		}
		root.Projects = []pmxmlProject{single}
	}

	for _, proj := range root.Projects {
		wbsByID := make(map[string]string, len(proj.WBSNodes))
		for _, w := range proj.WBSNodes {
			wbsByID[w.ObjectId] = w.Code
		}

		for i, act := range proj.Activities {
			actID := act.Id
			if actID == "" {
				actID = act.ObjectId
			}
			if actID == "" {
				result.Errors = append(result.Errors, SyncErrorEntry{
					Code:    "E001",
					Message: "missing activity Id / ObjectId",
					Row:     i + 1,
				})
				continue
			}

			baselinePhys := clamp100(act.PhysicalPercentComplete)
			actualPhys := clamp100(act.ActualPhysicalPercentComplete)
			if actualPhys == 0 {
				actualPhys = baselinePhys
			}

			raw := map[string]string{
				"ObjectId":                act.ObjectId,
				"Id":                      act.Id,
				"Name":                    act.Name,
				"WBSObjectId":             act.WBSObjectId,
				"PlannedStartDate":        act.PlannedStartDate,
				"PlannedFinishDate":       act.PlannedFinishDate,
				"ActualStartDate":         act.ActualStartDate,
				"ActualFinishDate":        act.ActualFinishDate,
				"PhysicalPercentComplete": fmt.Sprintf("%g", act.PhysicalPercentComplete),
			}

			result.Activities = append(result.Activities, ParsedActivity{
				ActivityID:       actID,
				WBSCode:          wbsByID[act.WBSObjectId],
				ActivityName:     act.Name,
				PlannedStart:     normalisePMXMLDate(act.PlannedStartDate),
				PlannedEnd:       normalisePMXMLDate(act.PlannedFinishDate),
				ActualStart:      normalisePMXMLDate(act.ActualStartDate),
				ActualEnd:        normalisePMXMLDate(act.ActualFinishDate),
				BaselinePhysical: baselinePhys,
				ActualPhysical:   actualPhys,
				RawPayload:       raw,
			})
		}
	}
	return result, nil
}

// normalisePMXMLDate converts PMXML ISO datetime strings → "2006-01-02".
func normalisePMXMLDate(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	for _, layout := range []string{
		time.RFC3339, "2006-01-02T15:04:05", "2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05", "2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return s
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func clamp100(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

// RawPayloadJSON serialises a map[string]string to a compact JSON string,
// used when storing raw_payload in the DB.
func RawPayloadJSON(m map[string]string) string {
	b, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(b)
}
