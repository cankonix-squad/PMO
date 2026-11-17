package kpbu

import (
	"testing"
	"time"

	"github.com/harmanto-49/cankora/internal/shared/types"
)

func ft(s string) *types.FlexTime {
	t, _ := time.Parse("2006-01-02", s)
	return &types.FlexTime{Time: t}
}

func TestFinancialCloseStatus(t *testing.T) {
	now, _ := time.Parse("2006-01-02", "2026-11-17")
	cases := []struct {
		name   string
		signed *types.FlexTime
		actual *types.FlexTime
		want   string
		days   *int
	}{
		{"not signed", nil, nil, "NONE", nil},
		{"on track", ft("2026-06-01"), nil, "ON_TRACK", intp(196)},
		{"warning", ft("2026-01-10"), nil, "WARNING", intp(54)},
		{"overdue", ft("2025-10-01"), nil, "OVERDUE", intp(-47)},
		{"achieved", ft("2025-10-01"), ft("2026-05-01"), "ACHIEVED", nil},
	}
	for _, c := range cases {
		p := &Project{AgreementSignedDate: c.signed, FinancialCloseActual: c.actual}
		financialCloseStatus(p, now)
		if p.FinancialCloseStatus != c.want {
			t.Errorf("%s: status %s, want %s", c.name, p.FinancialCloseStatus, c.want)
		}
		if (c.days == nil) != (p.FinancialCloseDaysLeft == nil) || (c.days != nil && *c.days != *p.FinancialCloseDaysLeft) {
			t.Errorf("%s: days %v, want %v", c.name, p.FinancialCloseDaysLeft, c.days)
		}
	}
}

func intp(i int) *int { return &i }

func TestItemStatus(t *testing.T) {
	if itemStatus(nil) != DocMissing {
		t.Error("empty should be MISSING")
	}
	// newest first: latest rejected but an older approved exists → APPROVED
	if itemStatus([]Document{{Status: DocRejected}, {Status: DocApproved}}) != DocApproved {
		t.Error("any approved should win")
	}
	if itemStatus([]Document{{Status: DocSubmitted}, {Status: DocRejected}}) != DocSubmitted {
		t.Error("latest status expected")
	}
}

func TestAppliesTo(t *testing.T) {
	all := DocumentRequirement{AppliesTo: AppliesAll}
	uns := DocumentRequirement{AppliesTo: AppliesUnsolicited}
	if !appliesTo(all, "SOLICITED") || !appliesTo(uns, "UNSOLICITED") || appliesTo(uns, "SOLICITED") {
		t.Error("appliesTo mismatch")
	}
}

func TestNormalizeRequirement(t *testing.T) {
	bad := "FOO"
	r := RequirementRequest{Code: "x1", Category: "penawaran", Name: "n", RequiredBeforeStage: &bad}
	if normalizeRequirement(&r) == nil {
		t.Error("invalid stage must fail")
	}
	st := " contract_signed "
	r = RequirementRequest{Code: "x1", Category: "penawaran", Name: "n", RequiredBeforeStage: &st}
	if err := normalizeRequirement(&r); err != nil || r.Code != "X1" || *r.RequiredBeforeStage != StageContractSigned || r.AppliesTo != AppliesAll || r.SourceStatus != SourceUnverified {
		t.Errorf("normalize failed: %v %+v", err, r)
	}
}

func TestDefaultRequirementsValid(t *testing.T) {
	seen := map[string]bool{}
	for _, d := range DefaultRequirements {
		if seen[d.Code] {
			t.Errorf("duplicate code %s", d.Code)
		}
		seen[d.Code] = true
		if !validCategories[d.Category] || !validAppliesTo[d.AppliesTo] || !validSources[d.Source] {
			t.Errorf("invalid enum in %s", d.Code)
		}
		if d.BeforeStage != "" && !isMainStage(d.BeforeStage) && d.BeforeStage != StageClosed {
			t.Errorf("invalid stage in %s", d.Code)
		}
	}
}
