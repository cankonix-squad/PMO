package benefit

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/audit"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("benefit resource not found")

type Service struct {
	db    *gorm.DB
	audit *audit.Writer
}

func NewService(db *gorm.DB, a *audit.Writer) *Service { return &Service{db: db, audit: a} }
func (s *Service) ListIndicators(org uuid.UUID) ([]Indicator, error) {
	var v []Indicator
	err := s.db.Where("organization_id = ?", org).Order("name").Find(&v).Error
	return v, err
}

func (s *Service) GetIndicator(org, id uuid.UUID) (*Indicator, error) {
	var v Indicator
	if err := s.db.Where("organization_id = ? AND id = ?", org, id).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &v, nil
}

func (s *Service) CreateIndicator(org, actor uuid.UUID, r CreateIndicatorRequest) (*Indicator, error) {
	if err := s.validateProject(org, r.ProjectID); err != nil {
		return nil, err
	}
	if err := s.validateUser(org, r.OwnerID); err != nil {
		return nil, err
	}
	now := time.Now()
	v := &Indicator{ID: uuid.New(), OrganizationID: org, ProjectID: r.ProjectID, Name: strings.TrimSpace(r.Name), Unit: normalize(r.Unit), AggregationMethod: normalize(r.AggregationMethod), OwnerID: r.OwnerID, Source: strings.TrimSpace(r.Source), Description: strings.TrimSpace(r.Description), CreatedBy: actor, CreatedAt: now, UpdatedAt: now}
	if err := s.db.Create(v).Error; err != nil {
		return nil, err
	}
	s.record(org, actor, "benefit_indicator.created", "benefit_indicator", v.ID.String())
	return v, nil
}

func (s *Service) UpdateIndicator(org, actor, id uuid.UUID, r UpdateIndicatorRequest) (*Indicator, error) {
	v, err := s.GetIndicator(org, id)
	if err != nil {
		return nil, err
	}
	if r.ProjectID != nil {
		if err := s.validateProject(org, r.ProjectID); err != nil {
			return nil, err
		}
		v.ProjectID = r.ProjectID
	}
	if r.OwnerID != nil {
		if err := s.validateUser(org, r.OwnerID); err != nil {
			return nil, err
		}
		v.OwnerID = r.OwnerID
	}
	if r.Name != "" {
		v.Name = strings.TrimSpace(r.Name)
	}
	if r.Unit != "" {
		v.Unit = normalize(r.Unit)
	}
	if r.AggregationMethod != "" {
		v.AggregationMethod = normalize(r.AggregationMethod)
	}
	if r.Source != "" {
		v.Source = strings.TrimSpace(r.Source)
	}
	if r.Description != "" {
		v.Description = strings.TrimSpace(r.Description)
	}
	v.UpdatedAt = time.Now()
	if err := s.db.Save(v).Error; err != nil {
		return nil, err
	}
	s.record(org, actor, "benefit_indicator.updated", "benefit_indicator", v.ID.String())
	return v, nil
}

func (s *Service) DeleteIndicator(org, actor, id uuid.UUID) error {
	v, err := s.GetIndicator(org, id)
	if err != nil {
		return err
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("organization_id = ? AND indicator_id = ?", org, id).Delete(&Measurement{}).Error; err != nil {
			return err
		}
		return tx.Delete(v).Error
	}); err != nil {
		return err
	}
	s.record(org, actor, "benefit_indicator.deleted", "benefit_indicator", id.String())
	return nil
}

func (s *Service) ListMeasurements(org, id uuid.UUID) ([]Measurement, error) {
	if _, err := s.GetIndicator(org, id); err != nil {
		return nil, err
	}
	var v []Measurement
	err := s.db.Where("organization_id = ? AND indicator_id = ?", org, id).Order("period_year DESC, period_month DESC").Find(&v).Error
	return v, err
}

func (s *Service) CreateMeasurement(org, actor, id uuid.UUID, r CreateMeasurementRequest) (*Measurement, error) {
	if _, err := s.GetIndicator(org, id); err != nil {
		return nil, err
	}
	if r.ValidationStatus == "" {
		r.ValidationStatus = "DRAFT"
	}
	now := time.Now()
	v := &Measurement{ID: uuid.New(), OrganizationID: org, IndicatorID: id, PeriodYear: r.PeriodYear, PeriodMonth: r.PeriodMonth, Baseline: r.Baseline, Target: r.Target, Actual: r.Actual, Source: strings.TrimSpace(r.Source), ValidationStatus: normalize(r.ValidationStatus), CreatedBy: actor, CreatedAt: now, UpdatedAt: now}
	if err := s.db.Create(v).Error; err != nil {
		return nil, err
	}
	s.record(org, actor, "benefit_measurement.created", "benefit_measurement", v.ID.String())
	return v, nil
}

func (s *Service) UpdateMeasurement(org, actor, indicatorID, measurementID uuid.UUID, r UpdateMeasurementRequest) (*Measurement, error) {
	if _, err := s.GetIndicator(org, indicatorID); err != nil {
		return nil, err
	}
	var v Measurement
	if err := s.db.Where("organization_id = ? AND indicator_id = ? AND id = ?", org, indicatorID, measurementID).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if r.PeriodYear != nil {
		v.PeriodYear = *r.PeriodYear
	}
	if r.PeriodMonth != nil {
		v.PeriodMonth = *r.PeriodMonth
	}
	if r.Baseline != nil {
		v.Baseline = *r.Baseline
	}
	if r.Target != nil {
		v.Target = *r.Target
	}
	if r.Actual != nil {
		v.Actual = *r.Actual
	}
	if r.Source != "" {
		v.Source = strings.TrimSpace(r.Source)
	}
	if r.ValidationStatus != "" {
		v.ValidationStatus = normalize(r.ValidationStatus)
	}
	v.UpdatedAt = time.Now()
	if err := s.db.Save(&v).Error; err != nil {
		return nil, err
	}
	s.record(org, actor, "benefit_measurement.updated", "benefit_measurement", v.ID.String())
	return &v, nil
}

func (s *Service) DeleteMeasurement(org, actor, indicatorID, measurementID uuid.UUID) error {
	if _, err := s.GetIndicator(org, indicatorID); err != nil {
		return err
	}
	result := s.db.Where("organization_id = ? AND indicator_id = ? AND id = ?", org, indicatorID, measurementID).Delete(&Measurement{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	s.record(org, actor, "benefit_measurement.deleted", "benefit_measurement", measurementID.String())
	return nil
}

func (s *Service) Aggregate(org, id uuid.UUID) (*AggregateResult, error) {
	i, err := s.GetIndicator(org, id)
	if err != nil {
		return nil, err
	}
	var v []Measurement
	if err := s.db.Where("organization_id = ? AND indicator_id = ? AND validation_status = ?", org, id, "VALID").Order("period_year DESC, period_month DESC, created_at DESC").Find(&v).Error; err != nil {
		return nil, err
	}
	if len(v) == 0 {
		return &AggregateResult{Indicator: *i, Count: 0, Value: nil}, nil
	}
	value := v[0].Actual
	switch i.AggregationMethod {
	case "SUM":
		value = 0
		for _, m := range v {
			value += m.Actual
		}
	case "AVERAGE":
		value = 0
		for _, m := range v {
			value += m.Actual
		}
		value /= float64(len(v))
	}
	return &AggregateResult{Indicator: *i, Count: len(v), Value: &value}, nil
}

func (s *Service) Summary(org uuid.UUID) ([]SummaryItem, error) {
	indicators, err := s.ListIndicators(org)
	if err != nil {
		return nil, err
	}
	grouped := map[string]SummaryItem{}
	for _, indicator := range indicators {
		aggregate, err := s.Aggregate(org, indicator.ID)
		if err != nil || aggregate.Value == nil {
			if err != nil {
				return nil, err
			}
			continue
		}
		key := indicator.Unit + "|" + indicator.AggregationMethod
		item := grouped[key]
		item.Unit = indicator.Unit
		item.AggregationMethod = indicator.AggregationMethod
		item.Count++
		item.Value += *aggregate.Value
		grouped[key] = item
	}
	result := make([]SummaryItem, 0, len(grouped))
	for _, item := range grouped {
		if item.AggregationMethod == "AVERAGE" && item.Count > 0 {
			item.Value = item.Value / float64(item.Count)
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *Service) validateProject(org uuid.UUID, projectID *uuid.UUID) error {
	if projectID == nil {
		return nil
	}
	var n int64
	if err := s.db.Table("projects").Where("id = ? AND organization_id = ? AND deleted_at IS NULL", *projectID, org).Count(&n).Error; err != nil {
		return err
	}
	if n != 1 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) validateUser(org uuid.UUID, userID *uuid.UUID) error {
	if userID == nil {
		return nil
	}
	var n int64
	if err := s.db.Table("users").Where("id = ? AND organization_id = ? AND deleted_at IS NULL", *userID, org).Count(&n).Error; err != nil {
		return err
	}
	if n != 1 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) record(org, actor uuid.UUID, action, entityType, entityID string) {
	if s.audit != nil {
		s.audit.Record(audit.WriteRequest{OrganizationID: org, ActorID: &actor, Action: action, EntityType: entityType, EntityID: entityID})
	}
}

func normalize(v string) string {
	return strings.ToUpper(strings.TrimSpace(v))
}
