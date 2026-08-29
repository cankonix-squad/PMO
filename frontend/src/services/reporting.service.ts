// P2-009 Reporting Analytics service

import api from '@/lib/axios'
import { ApiResponse } from '@/types/api'
import type {
  ReportDefinition,
  DatasetFilter,
  ExecutiveSummaryData,
  ProjectPerformanceRow,
  RiskIssueRow,
  BudgetRow,
  BenefitRow,
  PriorityRow,
  PowerBIConfig,
  ReportExportRequest,
  CreateExportRequestInput,
} from '@/types/reporting'

const BASE = '/analytics/reports'

function toParams(f?: DatasetFilter): Record<string, string> {
  const p: Record<string, string> = {}
  if (f?.period_start) p.period_start = f.period_start
  if (f?.period_end)   p.period_end   = f.period_end
  if (f?.program_id)   p.program_id   = f.program_id
  if (f?.status)       p.status       = f.status
  if (f?.province)     p.province     = f.province
  return p
}

export const reportingService = {
  // Catalog
  getCatalog: async (): Promise<ReportDefinition[]> => {
    const res = await api.get<ApiResponse<ReportDefinition[]>>(`${BASE}/catalog`)
    return res.data.data ?? []
  },

  // Datasets
  getExecutiveSummary: async (filter?: DatasetFilter): Promise<ExecutiveSummaryData> => {
    const res = await api.get<ApiResponse<ExecutiveSummaryData>>(
      `${BASE}/datasets/executive-summary`,
      { params: toParams(filter) },
    )
    return res.data.data
  },

  getProjectPerformance: async (filter?: DatasetFilter): Promise<ProjectPerformanceRow[]> => {
    const res = await api.get<ApiResponse<ProjectPerformanceRow[]>>(
      `${BASE}/datasets/project-performance`,
      { params: toParams(filter) },
    )
    return res.data.data ?? []
  },

  getRiskIssue: async (filter?: DatasetFilter): Promise<RiskIssueRow[]> => {
    const res = await api.get<ApiResponse<RiskIssueRow[]>>(
      `${BASE}/datasets/risk-issue`,
      { params: toParams(filter) },
    )
    return res.data.data ?? []
  },

  getBudget: async (filter?: DatasetFilter): Promise<BudgetRow[]> => {
    const res = await api.get<ApiResponse<BudgetRow[]>>(
      `${BASE}/datasets/budget`,
      { params: toParams(filter) },
    )
    return res.data.data ?? []
  },

  getBenefits: async (filter?: DatasetFilter): Promise<BenefitRow[]> => {
    const res = await api.get<ApiResponse<BenefitRow[]>>(
      `${BASE}/datasets/benefits`,
      { params: toParams(filter) },
    )
    return res.data.data ?? []
  },

  getPriority: async (): Promise<PriorityRow[]> => {
    const res = await api.get<ApiResponse<PriorityRow[]>>(`${BASE}/datasets/priority`)
    return res.data.data ?? []
  },

  // Power BI config
  getPowerBIConfig: async (): Promise<PowerBIConfig> => {
    const res = await api.get<ApiResponse<PowerBIConfig>>(`${BASE}/powerbi/config`)
    return res.data.data
  },

  // Export requests
  createExportRequest: async (input: CreateExportRequestInput): Promise<ReportExportRequest> => {
    const res = await api.post<{ data: ReportExportRequest }>(`${BASE}/export/request`, input)
    return res.data.data
  },

  listExportRequests: async (): Promise<ReportExportRequest[]> => {
    const res = await api.get<ApiResponse<ReportExportRequest[]>>(`${BASE}/export/requests`)
    return res.data.data ?? []
  },

  getExportRequest: async (id: string): Promise<ReportExportRequest> => {
    const res = await api.get<ApiResponse<ReportExportRequest>>(`${BASE}/export/requests/${id}`)
    return res.data.data
  },

  // Download completed export file — returns a blob URL for the browser to download.
  // Uses authenticated axios so the Authorization header is sent.
  downloadExportFile: async (id: string, fileName: string): Promise<void> => {
    const res = await api.get(`${BASE}/export/requests/${id}/download`, {
      responseType: 'blob',
    })
    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', fileName || `export-${id}.bin`)
    document.body.appendChild(link)
    link.click()
    link.parentNode?.removeChild(link)
    window.URL.revokeObjectURL(url)
  },
}
