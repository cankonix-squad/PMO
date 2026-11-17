package kpbu

import (
	"testing"
	"time"
)

func TestEnrichPayment(t *testing.T) {
	now, _ := time.Parse("2006-01-02", "2026-06-15")
	cases := []struct {
		name string
		p    AvailabilityPayment
		want string
		out  float64
	}{
		{"scheduled future", AvailabilityPayment{ScheduledAmount: 100, DueDate: ft("2026-07-01")}, "SCHEDULED", 100},
		{"overdue", AvailabilityPayment{ScheduledAmount: 100, PaidAmount: 40, DueDate: ft("2026-06-01")}, "OVERDUE", 60},
		{"partial not due", AvailabilityPayment{ScheduledAmount: 100, PaidAmount: 40, DueDate: ft("2026-07-01")}, "PARTIAL", 60},
		{"paid net after deduction", AvailabilityPayment{ScheduledAmount: 100, DeductionAmount: 10, PaidAmount: 90, DueDate: ft("2026-06-01")}, "PAID", 0},
		{"fully deducted with paid date", AvailabilityPayment{ScheduledAmount: 50, DeductionAmount: 50, PaidDate: ft("2026-06-02")}, "PAID", 0},
	}
	for _, c := range cases {
		p := c.p
		enrichPayment(&p, now)
		if p.Status != c.want || p.Outstanding != c.out {
			t.Errorf("%s: got %s/%v want %s/%v", c.name, p.Status, p.Outstanding, c.want, c.out)
		}
	}
}

func TestNormalizePayment(t *testing.T) {
	if err := normalizePayment(&AvailabilityPaymentRequest{ScheduledAmount: 100, DeductionAmount: 10}); err == nil {
		t.Fatal("deduction without reason must fail")
	}
	if err := normalizePayment(&AvailabilityPaymentRequest{ScheduledAmount: 100, DeductionAmount: 10, DeductionReason: "KPI", PaidAmount: 95}); err == nil {
		t.Fatal("paid above net must fail")
	}
	if err := normalizePayment(&AvailabilityPaymentRequest{ScheduledAmount: 100, DeductionAmount: 10, DeductionReason: "KPI", PaidAmount: 90}); err != nil {
		t.Fatal(err)
	}
}

func TestBuildFinanceSummary(t *testing.T) {
	p := &Project{Currency: "IDR", InvestmentValue: 1000}
	fin := []Financing{
		{Source: "EQUITY", Amount: 300, DisbursedAmount: 300, Currency: "IDR", Status: "DISBURSED"},
		{Source: "SENIOR_DEBT", Amount: 600, Currency: "IDR", Status: "COMMITTED"},
		{Source: "SENIOR_DEBT", Amount: 999, Currency: "IDR", Status: "CANCELLED"},
		{Source: "BOND", Amount: 50, Currency: "USD", Status: "PLANNED"},
	}
	sup := []GovernmentSupport{{Type: "VGF", Amount: 100, Currency: "IDR", Status: "APPROVED"}, {Type: "PERMIT", Status: "PROPOSED", Currency: "IDR"}}
	s := buildFinanceSummary(p, fin, sup, nil)
	if s.FinancingTotal != 900 || s.EquityTotal != 300 || s.DebtTotal != 600 {
		t.Fatalf("totals: %+v", s)
	}
	if *s.EquityPct != 33.33 || *s.DebtToEquity != 2 || *s.FundingCoveragePct != 90 {
		t.Fatalf("ratios: %v %v %v", *s.EquityPct, *s.DebtToEquity, *s.FundingCoveragePct)
	}
	if s.SkippedOtherCcy != 1 || s.SupportApproved != 100 || s.SupportCount != 2 {
		t.Fatalf("support/ccy: %+v", s)
	}
}

func TestNormalizeFinancing(t *testing.T) {
	if err := normalizeFinancing(&FinancingRequest{Source: "equity", Amount: 10, DisbursedAmount: 20}); err == nil {
		t.Fatal("disbursed above amount must fail")
	}
	r := FinancingRequest{Source: "equity", Amount: 10}
	if err := normalizeFinancing(&r); err != nil || r.Source != "EQUITY" || r.Status != "PLANNED" || r.Currency != "IDR" {
		t.Fatalf("%v %+v", err, r)
	}
}
