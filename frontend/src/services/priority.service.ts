import { api } from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type {
  BatchCalculateRequest,
  BatchCalculateResponse,
  CalculateRequest,
  CreateFormulaRequest,
  Formula,
  RankingResponse,
  Score,
  UpdateFormulaRequest,
} from "@/types/priority";

export const priorityService = {
  // ── Formulas ──────────────────────────────────────────────────────────────

  listFormulas: async (): Promise<Formula[]> => {
    const response = await api.get<ApiResponse<Formula[]>>("/priority/formulas");
    return response.data.data;
  },

  getFormula: async (id: string): Promise<Formula> => {
    const response = await api.get<ApiResponse<Formula>>(`/priority/formulas/${id}`);
    return response.data.data;
  },

  createFormula: async (payload: CreateFormulaRequest): Promise<Formula> => {
    const response = await api.post<ApiResponse<Formula>>("/priority/formulas", payload);
    return response.data.data;
  },

  updateFormula: async (id: string, payload: UpdateFormulaRequest): Promise<Formula> => {
    const response = await api.put<ApiResponse<Formula>>(`/priority/formulas/${id}`, payload);
    return response.data.data;
  },

  activateFormula: async (id: string): Promise<Formula> => {
    const response = await api.post<ApiResponse<Formula>>(`/priority/formulas/${id}/activate`, {});
    return response.data.data;
  },

  // ── Scoring ───────────────────────────────────────────────────────────────

  calculate: async (payload: CalculateRequest): Promise<Score> => {
    const response = await api.post<ApiResponse<Score>>("/priority/calculate", payload);
    return response.data.data;
  },

  batchCalculate: async (payload: BatchCalculateRequest): Promise<BatchCalculateResponse> => {
    const response = await api.post<ApiResponse<BatchCalculateResponse>>(
      "/priority/batch-calculate",
      payload
    );
    return response.data.data;
  },

  // ── Ranking & Scores ──────────────────────────────────────────────────────

  listRanking: async (formulaId?: string): Promise<RankingResponse> => {
    const params = formulaId ? { formula_id: formulaId } : {};
    const response = await api.get<ApiResponse<RankingResponse>>("/priority/projects", { params });
    return response.data.data;
  },

  getProjectScore: async (projectId: string): Promise<Score> => {
    const response = await api.get<ApiResponse<Score>>(`/priority/projects/${projectId}`);
    return response.data.data;
  },

  explainProjectScore: async (projectId: string): Promise<Score> => {
    const response = await api.get<ApiResponse<Score>>(`/priority/projects/${projectId}/explain`);
    return response.data.data;
  },
};
