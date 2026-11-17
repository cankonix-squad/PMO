package kpbu

import (
	"testing"

	"github.com/google/uuid"
)

func TestKPIMet(t *testing.T) {
	hi := KPI{TargetValue: 95, Direction: KPIHigherBetter}
	lo := KPI{TargetValue: 2, Direction: KPILowerBetter}
	if !kpiMet(hi, 95) || kpiMet(hi, 94.9) || !kpiMet(lo, 2) || kpiMet(lo, 2.1) {
		t.Fatal("kpiMet direction handling is wrong")
	}
}

func TestBuildScorecard(t *testing.T) {
	a := KPI{ID: uuid.New(), TargetValue: 95, Direction: KPIHigherBetter, DeductionPct: 5}
	b := KPI{ID: uuid.New(), TargetValue: 2, Direction: KPILowerBetter, DeductionPct: 70}
	c := KPI{ID: uuid.New(), TargetValue: 1, DeductionPct: 50}
	d := KPI{ID: uuid.New(), TargetValue: 1, DeductionPct: 50}
	e := KPI{ID: uuid.New(), TargetValue: 1, Direction: KPIHigherBetter, DeductionPct: 40}
	ms := []KPIMeasurement{
		{KPIID: a.ID, ActualValue: 90, Status: "VERIFIED"},  // missed 5
		{KPIID: b.ID, ActualValue: 3, Status: "VERIFIED"},   // missed 70
		{KPIID: c.ID, ActualValue: 0, Status: "SUBMITTED"},  // pending, no deduction
		{KPIID: d.ID, ActualValue: 0, Status: "REJECTED"},   // ignored -> no data
		{KPIID: e.ID, ActualValue: 0.5, Status: "VERIFIED"}, // missed 40 -> capped
	}
	sc := buildScorecard(2026, 5, []KPI{a, b, c, d, e}, ms)
	if sc.Missed != 3 || sc.Pending != 1 || sc.NoData != 1 || sc.Met != 0 {
		t.Fatalf("counts: %+v", sc)
	}
	if sc.SuggestedDeductionPct != 100 {
		t.Fatalf("deduction must be capped at 100, got %v", sc.SuggestedDeductionPct)
	}
}

func TestNormalizeKPI(t *testing.T) {
	r := KPIRequest{Code: " ab-1 ", Name: "Ketersediaan air"}
	if err := normalizeKPI(&r); err != nil || r.Code != "AB-1" || r.Direction != KPIHigherBetter {
		t.Fatalf("%v %+v", err, r)
	}
	if err := normalizeKPI(&KPIRequest{Code: "X", Name: "Y", Direction: "UP"}); err == nil {
		t.Fatal("invalid direction must fail")
	}
}
