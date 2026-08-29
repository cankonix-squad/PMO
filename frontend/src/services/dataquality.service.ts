import { api } from '@/lib/axios';
import type {
  CreateValidationSubmissionRequest,
  TransitionValidationRequest,
  ValidationStatus,
  ValidationSubmission,
} from '@/types/dataquality';

interface ApiResponse<T> {
  message: string;
  data: T;
}

export const dataQualityService = {
  list: async (status?: ValidationStatus): Promise<ValidationSubmission[]> => {
    const response = await api.get<ApiResponse<ValidationSubmission[]>>('/validation-queue', {
      params: status ? { status } : undefined,
    });
    return response.data.data;
  },

  submit: async (data: CreateValidationSubmissionRequest): Promise<ValidationSubmission> => {
    const response = await api.post<ApiResponse<ValidationSubmission>>('/validation-queue', data);
    return response.data.data;
  },

  transition: async (id: string, data: TransitionValidationRequest): Promise<ValidationSubmission> => {
    const response = await api.patch<ApiResponse<ValidationSubmission>>(
      `/validation-queue/${id}/status`,
      data
    );
    return response.data.data;
  },
};