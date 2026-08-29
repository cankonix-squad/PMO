export type HealthClass = 'GREEN' | 'YELLOW' | 'RED' | 'CRITICAL';

export interface HealthSnapshot {
  id: string;
  project_id: string;
  formula_id: string;
  period_year: number;
  period_month: number;
  score: number;
  health_class: HealthClass;
  components: Record<string, { score: number; weight: number; available: boolean; reason: string }>;
  explanation: string;
  calculated_at: string;
}

export interface CalculateHealthRequest {
  formula_id?: string;
  period_year: number;
  period_month: number;
}