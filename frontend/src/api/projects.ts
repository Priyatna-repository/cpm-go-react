import { apiClient } from './client';

export interface ProjectCompanyRef {
  id: number;
  name: string;
}

export interface ProjectUserRef {
  id: number;
  name: string;
}

export interface ProjectLabelRef {
  id: number;
  name: string;
  slug: string;
  color?: string;
}

export interface Project {
  id: number;
  code: string;
  name: string;
  description?: string;
  client_company?: ProjectCompanyRef;
  client_user?: ProjectUserRef;
  start_date?: string;
  end_date?: string;
  budget_estimate?: number;
  type_label?: ProjectLabelRef;
  status_labels: ProjectLabelRef[];
  is_completed: boolean;
  completed_at?: string;
  archived_at?: string;
  users: ProjectUserRef[];
}

export interface ProjectListResponse {
  data: Project[];
  meta: { current_page: number; last_page: number; total: number };
}

export interface ListProjectsParams {
  page?: number;
  search?: string;
  archived?: boolean;
}

export async function listProjects(params: ListProjectsParams = {}): Promise<ProjectListResponse> {
  const { data } = await apiClient.get<ProjectListResponse>('/api/v1/projects', {
    params: {
      page: params.page,
      search: params.search || undefined,
      archived: params.archived ? '1' : undefined,
    },
  });
  return data;
}

export interface ProjectInput {
  name: string;
  description?: string;
  client_company_id?: number;
  client_user_id?: number;
  start_date?: string;
  end_date?: string;
  budget_estimate?: number;
  type_label_id?: number;
  status_label_ids: number[];
}

export async function createProject(input: ProjectInput): Promise<Project> {
  const { data } = await apiClient.post<Project>('/api/v1/projects', input);
  return data;
}

export async function updateProject(id: number, input: ProjectInput): Promise<Project> {
  const { data } = await apiClient.put<Project>(`/api/v1/projects/${id}`, input);
  return data;
}

export async function archiveProject(id: number): Promise<void> {
  await apiClient.delete(`/api/v1/projects/${id}`);
}

export async function restoreProject(id: number): Promise<void> {
  await apiClient.post(`/api/v1/projects/${id}/restore`);
}

export async function forceDeleteProject(id: number): Promise<void> {
  await apiClient.delete(`/api/v1/projects/${id}/force`);
}

export async function updateProjectAccess(id: number, userIds: number[]): Promise<Project> {
  const { data } = await apiClient.put<Project>(`/api/v1/projects/${id}/access`, { user_ids: userIds });
  return data;
}
