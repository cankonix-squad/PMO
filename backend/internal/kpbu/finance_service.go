package kpbu

import (
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ── Generic child-record helpers (scoped by org + KPBU project) ─────────────

type financeRecord interface {
	Financing | GovernmentSupport | AvailabilityPayment
}

func listChildren[T financeRecord](s *Service, orgID, projectID uuid.UUID, order string) ([]T, error) {
	if _, err := s.findProject(orgID, projectID); err != nil {
		return nil, err
	}
	var list []T
	err := s.db.Where("organization_id = ? AND kpbu_project_id = ?", orgID, projectID).Order(order).Find(&list).Error
	return list, err
}

func findChild[T financeRecord](s *Service, orgID, projectID, id uuid.UUID) (*T, error) {
	var v T
	err := s.db.Where("id = ? AND organization_id = ? AND kpbu_project_id = ?", id, orgID, projectID).First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func deleteChild[T financeRecord](s *Service, orgID, projectID, id uuid.UUID) (*T, error) {
	v, err := findChild[T](s, orgID, projectID, id)
	if err != nil {
		return nil, err
	}
	return v, s.db.Delete(v).Error
}

func upperOr(v, def string) string {
	v = strings.ToUpper(strings.TrimSpace(v))
	if v == "" {
		return def
	}
	return v
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

// ── Financing ───────────────────────────────────────────────────────────────

func normalizeFinancing(r *FinancingRequest) error {
	r.Source = upperOr(r.Source, "")
	r.Status = upperOr(r.Status, "PLANNED")
	r.Currency = upperOr(r.Currency, "IDR")
	if !validFinancingSources[r.Source] {
		return invalid("invalid financing source")
	}
	if !validFinancingStatus[r.Status] {
		return invalid("invalid financing status")
	}
	if r.DisbursedAmount > r.Amount {
		return invalid("disbursed_amount must not exceed amount")
	}
	return nil
}

func applyFinancing(f *Financing, r FinancingRequest) {
	f.Source, f.Provider, f.AgreementNumber = r.Source, strings.TrimSpace(r.Provider), strings.TrimSpace(r.AgreementNumber)
	f.AgreementDate, f.Amount, f.DisbursedAmount = r.AgreementDate, r.Amount, r.DisbursedAmount
	f.Currency, f.Status, f.Notes = r.Currency, r.Status, strings.TrimSpace(r.Notes)
}

func (s *Service) ListFinancings(orgID, projectID uuid.UUID) ([]Financing, error) {
	return listChildren[Financing](s, orgID, projectID, "created_at ASC")
}

func (s *Service) SaveFinancing(orgID, projectID uuid.UUID, id *uuid.UUID, r FinancingRequest) (*Financing, error) {
	if err := normalizeFinancing(&r); err != nil {
		return nil, err
	}
	if _, err := s.findProject(orgID, projectID); err != nil {
		return nil, err
	}
	f := &Financing{OrganizationID: orgID, KPBUProjectID: projectID}
	if id != nil {
		var err error
		if f, err = findChild[Financing](s, orgID, projectID, *id); err != nil {
			return nil, err
		}
	}
	applyFinancing(f, r)
	return f, s.db.Save(f).Error
}

func (s *Service) DeleteFinancing(orgID, projectID, id uuid.UUID) (*Financing, error) {
	return deleteChild[Financing](s, orgID, projectID, id)
}

// ── Government support ──────────────────────────────────────────────────────

func normalizeSupport(r *GovernmentSupportRequest) error {
	r.Type = upperOr(r.Type, "")
	r.Status = upperOr(r.Status, "PROPOSED")
	r.Currency = upperOr(r.Currency, "IDR")
	if !validSupportTypes[r.Type] {
		return invalid("invalid government support type")
	}
	if !validSupportStatus[r.Status] {
		return invalid("invalid government support status")
	}
	return nil
}

func (s *Service) ListSupports(orgID, projectID uuid.UUID) ([]GovernmentSupport, error) {
	return listChildren[GovernmentSupport](s, orgID, projectID, "created_at ASC")
}

func (s *Service) SaveSupport(orgID, projectID uuid.UUID, id *uuid.UUID, r GovernmentSupportRequest) (*GovernmentSupport, error) {
	if err := normalizeSupport(&r); err != nil {
		return nil, err
	}
	if _, err := s.findProject(orgID, projectID); err != nil {
		return nil, err
	}
	g := &GovernmentSupport{OrganizationID: orgID, KPBUProjectID: projectID}
	if id != nil {
		var err error
		if g, err = findChild[GovernmentSupport](s, orgID, projectID, *id); err != nil {
			return nil, err
		}
	}
	g.Type, g.Provider, g.DecisionNumber = r.Type, strings.TrimSpace(r.Provider), strings.TrimSpace(r.DecisionNumber)
	g.DecisionDate, g.Amount, g.Currency, g.Status, g.Notes = r.DecisionDate, r.Amount, r.Currency, r.Status, strings.TrimSpace(r.Notes)
	return g, s.db.Save(g).Error
}

func (s *Service) DeleteSupport(orgID, projectID, id uuid.UUID) (*GovernmentSupport, error) {
	return deleteChild[GovernmentSupport](s, orgID, projectID, id)
}

// ── Availability payments ───────────────────────────────────────────────────

// enrichPayment computes net, outstanding and status.
//
//	PAID      paid >= net (and net > 0, or a paid date is set)
//	PARTIAL   0 < paid < net, not yet past due
//	OVERDUE   due date passed and paid < net
//	SCHEDULED otherwise
func enrichPayment(p *AvailabilityPayment, now time.Time) {
	p.NetAmount = round2(p.ScheduledAmount - p.DeductionAmount)
	p.Outstanding = round2(math.Max(0, p.NetAmount-p.PaidAmount))
	p.DaysOverdue = 0
	switch {
	case p.Outstanding == 0 && (p.NetAmount > 0 || p.PaidDate != nil):
		p.Status = "PAID"
	case p.DueDate != nil && !p.DueDate.IsZero() && today(p.DueDate.Time).Before(today(now)):
		p.Status = "OVERDUE"
		p.DaysOverdue = int(today(now).Sub(today(p.DueDate.Time)).Hours() / 24)
	case p.PaidAmount > 0:
		p.Status = "PARTIAL"
	default:
		p.Status = "SCHEDULED"
	}
}

func normalizePayment(r *AvailabilityPaymentRequest) error {
	r.Currency = upperOr(r.Currency, "IDR")
	if r.DeductionAmount > r.ScheduledAmount {
		return invalid("deduction_amount must not exceed scheduled_amount")
	}
	if r.PaidAmount > r.ScheduledAmount-r.DeductionAmount+0.005 {
		return invalid("paid_amount must not exceed the net amount (scheduled - deduction)")
	}
	if r.DeductionAmount > 0 && strings.TrimSpace(r.DeductionReason) == "" {
		return invalid("deduction_reason is required when a deduction is applied")
	}
	return nil
}

func (s *Service) ListPayments(orgID, projectID uuid.UUID) ([]AvailabilityPayment, error) {
	list, err := listChildren[AvailabilityPayment](s, orgID, projectID, "period_year ASC, period_month ASC")
	now := s.now()
	for i := range list {
		enrichPayment(&list[i], now)
	}
	return list, err
}

func (s *Service) SavePayment(orgID, projectID uuid.UUID, id *uuid.UUID, r AvailabilityPaymentRequest) (*AvailabilityPayment, error) {
	if err := normalizePayment(&r); err != nil {
		return nil, err
	}
	if _, err := s.findProject(orgID, projectID); err != nil {
		return nil, err
	}
	p := &AvailabilityPayment{OrganizationID: orgID, KPBUProjectID: projectID}
	if id != nil {
		var err error
		if p, err = findChild[AvailabilityPayment](s, orgID, projectID, *id); err != nil {
			return nil, err
		}
	}
	p.PeriodYear, p.PeriodMonth, p.DueDate = r.PeriodYear, r.PeriodMonth, r.DueDate
	p.ScheduledAmount, p.DeductionAmount, p.DeductionReason = r.ScheduledAmount, r.DeductionAmount, strings.TrimSpace(r.DeductionReason)
	p.PaidAmount, p.PaidDate, p.InvoiceNumber = r.PaidAmount, r.PaidDate, strings.TrimSpace(r.InvoiceNumber)
	p.Currency, p.Notes = r.Currency, strings.TrimSpace(r.Notes)
	if err := s.db.Save(p).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicatePeriod
		}
		return nil, err
	}
	enrichPayment(p, s.now())
	return p, nil
}

func (s *Service) DeletePayment(orgID, projectID, id uuid.UUID) (*AvailabilityPayment, error) {
	return deleteChild[AvailabilityPayment](s, orgID, projectID, id)
}

// ── Summary ─────────────────────────────────────────────────────────────────

func (s *Service) FinanceSummary(orgID, projectID uuid.UUID) (*FinanceSummary, error) {
	p, err := s.findProject(orgID, projectID)
	if err != nil {
		return nil, err
	}
	fin, err := s.ListFinancings(orgID, projectID)
	if err != nil {
		return nil, err
	}
	sup, err := s.ListSupports(orgID, projectID)
	if err != nil {
		return nil, err
	}
	pay, err := s.ListPayments(orgID, projectID)
	if err != nil {
		return nil, err
	}
	return buildFinanceSummary(p, fin, sup, pay), nil
}

func pctOf(a, b float64) *float64 {
	if b <= 0 {
		return nil
	}
	v := round2(a / b * 100)
	return &v
}

func buildFinanceSummary(p *Project, fin []Financing, sup []GovernmentSupport, pay []AvailabilityPayment) *FinanceSummary {
	ccy := p.Currency
	if ccy == "" {
		ccy = "IDR"
	}
	out := &FinanceSummary{Currency: ccy, InvestmentValue: p.InvestmentValue, FinancingBySource: map[string]float64{}}
	for _, f := range fin {
		if f.Status == "CANCELLED" {
			continue
		}
		if f.Currency != ccy {
			out.SkippedOtherCcy++
			continue
		}
		out.FinancingTotal += f.Amount
		out.FinancingBySource[f.Source] += f.Amount
		out.DisbursedTotal += f.DisbursedAmount
		if f.Source == "EQUITY" {
			out.EquityTotal += f.Amount
		} else if debtSources[f.Source] {
			out.DebtTotal += f.Amount
		}
	}
	out.EquityPct = pctOf(out.EquityTotal, out.FinancingTotal)
	if out.EquityTotal > 0 {
		v := round2(out.DebtTotal / out.EquityTotal)
		out.DebtToEquity = &v
	}
	out.FundingCoveragePct = pctOf(out.FinancingTotal, p.InvestmentValue)
	for _, g := range sup {
		out.SupportCount++
		if g.Currency != ccy {
			out.SkippedOtherCcy++
			continue
		}
		if supportCountsAsActive[g.Status] {
			out.SupportApproved += g.Amount
		} else if g.Status == "PROPOSED" || g.Status == "IN_PROCESS" {
			out.SupportProposed += g.Amount
		}
	}
	for _, a := range pay {
		if a.Currency != ccy {
			out.SkippedOtherCcy++
			continue
		}
		out.APScheduled += a.ScheduledAmount
		out.APDeductions += a.DeductionAmount
		out.APNet += a.NetAmount
		out.APPaid += a.PaidAmount
		if a.Status == "OVERDUE" {
			out.APOverdueCount++
			out.APOutstandingDue += a.Outstanding
		}
	}
	for k, v := range out.FinancingBySource {
		out.FinancingBySource[k] = round2(v)
	}
	out.FinancingTotal, out.DisbursedTotal = round2(out.FinancingTotal), round2(out.DisbursedTotal)
	out.EquityTotal, out.DebtTotal = round2(out.EquityTotal), round2(out.DebtTotal)
	out.SupportApproved, out.SupportProposed = round2(out.SupportApproved), round2(out.SupportProposed)
	out.APScheduled, out.APDeductions, out.APNet = round2(out.APScheduled), round2(out.APDeductions), round2(out.APNet)
	out.APPaid, out.APOutstandingDue = round2(out.APPaid), round2(out.APOutstandingDue)
	return out
}
