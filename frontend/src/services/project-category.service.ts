import api from "@/lib/axios";
import type { ApiResponse } from "@/types/api";
import type {
  ProjectCategory,
  CreateProjectCategoryRequest,
  UpdateProjectCategoryRequest,
} from "@/types/project-category";

export const projectCategoryService = {
  list(includeInactive = false): Promise<ApiResponse<ProjectCategory[]>> {
    return api
      .get("/project-categories", {
        params: includeInactive ? { include_inactive: "true" } : {},
      })
      .then((r) => r.data);
  },

  getById(id: string): Promise<ApiResponse<ProjectCategory>> {
    return api.get(`/project-categories/${id}`).then((r) => r.data);
  },

  create(payload: CreateProjectCategoryRequest): Promise<ApiResponse<ProjectCategory>> {
    return api.post("/project-categories", payload).then((r) => r.data);
  },

  update(id: string, payload: UpdateProjectCategoryRequest): Promise<ApiResponse<ProjectCategory>> {
    return api.put(`/project-categories/${id}`, payload).then((r) => r.data);
  },

  delete(id: string): Promise<void> {
    return api.delete(`/project-categories/${id}`).then(() => undefined);
  },
};
