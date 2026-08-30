import api from '@/lib/axios'
import { ApiResponse } from '@/types/api'
import { GISProjectMarker, GISSummary, GISFilter } from '@/types/gis'

export const gisService = {
  getProjects: async (filter?: GISFilter): Promise<GISProjectMarker[]> => {
    const params: Record<string, string> = {}
    if (filter?.status) params.status = filter.status
    if (filter?.health_class) params.health_class = filter.health_class
    if (filter?.province) params.province = filter.province

    const res = await api.get<ApiResponse<GISProjectMarker[]>>('/analytics/gis/projects', { params })
    return res.data.data ?? []
  },

  getSummary: async (): Promise<GISSummary> => {
    const res = await api.get<ApiResponse<GISSummary>>('/analytics/gis/summary')
    return res.data.data
  },

  getProjectDetail: async (projectId: string): Promise<GISProjectMarker> => {
    const res = await api.get<ApiResponse<GISProjectMarker>>(`/analytics/gis/projects/${projectId}`)
    return res.data.data
  },
}
