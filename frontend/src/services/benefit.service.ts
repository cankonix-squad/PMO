import { api } from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type {
  BenefitAggregate,
  BenefitIndicator,
  BenefitMeasurement,
  BenefitSummaryItem,
  CreateBenefitIndicatorRequest,
  CreateBenefitMeasurementRequest,
} from "@/types/benefit";

export const benefitService = {
  list: async (): Promise<BenefitIndicator[]> => {
    const response = await api.get<ApiResponse<BenefitIndicator[]>>("/benefits");
    return response.data.data;
  },

  create: async (payload: CreateBenefitIndicatorRequest): Promise<BenefitIndicator> => {
    const response = await api.post<ApiResponse<BenefitIndicator>>("/benefits", payload);
    return response.data.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/benefits/${id}`);
  },

  summary: async (): Promise<BenefitSummaryItem[]> => {
    const response = await api.get<ApiResponse<BenefitSummaryItem[]>>("/benefits/summary");
    return response.data.data;
  },

  aggregate: async (id: string): Promise<BenefitAggregate> => {
    const response = await api.get<ApiResponse<BenefitAggregate>>(`/benefits/${id}/aggregate`);
    return response.data.data;
  },

  listMeasurements: async (id: string): Promise<BenefitMeasurement[]> => {
    const response = await api.get<ApiResponse<BenefitMeasurement[]>>(`/benefits/${id}/measurements`);
    return response.data.data;
  },

  createMeasurement: async (
    id: string,
    payload: CreateBenefitMeasurementRequest
  ): Promise<BenefitMeasurement> => {
    const response = await api.post<ApiResponse<BenefitMeasurement>>(
      `/benefits/${id}/measurements`,
      payload
    );
    return response.data.data;
  },

  deleteMeasurement: async (indicatorID: string, measurementID: string): Promise<void> => {
    await api.delete(`/benefits/${indicatorID}/measurements/${measurementID}`);
  },
};
