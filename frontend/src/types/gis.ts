import { UUID } from 'crypto'

export interface GISProjectMarker {
  project_id: string
  project_code: string
  project_name: string
  status: string
  health_class: 'GREEN' | 'YELLOW' | 'RED' | 'CRITICAL' | 'UNSCORED'
  progress_pct: number
  budget_total: number
  priority_score: number
  latitude: number | null
  longitude: number | null
  province: string
  city: string
  location_name: string
  region_name: string
  open_risks: number
  open_issues: number
}

export interface GISSummary {
  total_projects: number
  mapped_projects: number
  unmapped_projects: number
  avg_progress_pct: number
  health_green: number
  health_yellow: number
  health_red: number
  health_critical: number
  health_unscored: number
}

export interface GISFilter {
  status?: string
  health_class?: string
  province?: string
}
