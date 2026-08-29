export type BenefitAggregationMethod = "SUM" | "AVERAGE" | "LATEST";
export type BenefitValidationStatus = "DRAFT" | "SUBMITTED" | "VALID" | "REJECTED" | "STALE";

export interface BenefitIndicator {
  id: string;
  organization_id: string;
  project_id?: string;
  name: string;
  unit: string;
  aggregation_method: BenefitAggregationMethod;
  owner_id?: string;
  source?: string;
  description?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface BenefitMeasurement {
  id: string;
  organization_id: string;
  indicator_id: string;
  period_year: number;
  period_month: number;
  baseline: number;
  target: number;
  actual: number;
  validation_status: BenefitValidationStatus;
  source?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface BenefitAggregate {
  indicator: BenefitIndicator;
  count: number;
  value: number | null;
}

export interface BenefitSummaryItem {
  unit: string;
  aggregation_method: BenefitAggregationMethod;
  count: number;
  value: number;
}

export interface CreateBenefitIndicatorRequest {
  project_id?: string;
  name: string;
  unit: string;
  aggregation_method: BenefitAggregationMethod;
  source?: string;
  description?: string;
}

export interface CreateBenefitMeasurementRequest {
  period_year: number;
  period_month: number;
  baseline: number;
  target: number;
  actual: number;
  source?: string;
  validation_status?: BenefitValidationStatus;
}
