import { apiClient } from './client';

export interface Label {
  id: number;
  name: string;
  slug: string;
  type: string;
  color?: string;
  icon?: string;
  archived_at?: string;
}

export interface LabelListResponse {
  data: Label[];
  meta: { current_page: number; last_page: number; total: number };
}

export interface ListLabelsParams {
  page?: number;
  search?: string;
  type?: string;
  archived?: boolean;
}

export async function listLabels(params: ListLabelsParams = {}): Promise<LabelListResponse> {
  const { data } = await apiClient.get<LabelListResponse>('/api/v1/labels', {
    params: {
      page: params.page,
      search: params.search || undefined,
      type: params.type || undefined,
      archived: params.archived ? '1' : undefined,
    },
  });
  return data;
}

export interface LabelInput {
  name: string;
  slug?: string;
  type: string;
  color: string;
  icon?: string;
}

export async function createLabel(input: LabelInput): Promise<Label> {
  const { data } = await apiClient.post<Label>('/api/v1/labels', input);
  return data;
}

export async function updateLabel(id: number, input: LabelInput): Promise<Label> {
  const { data } = await apiClient.put<Label>(`/api/v1/labels/${id}`, input);
  return data;
}

export async function archiveLabel(id: number): Promise<void> {
  await apiClient.delete(`/api/v1/labels/${id}`);
}

export async function restoreLabel(id: number): Promise<void> {
  await apiClient.post(`/api/v1/labels/${id}/restore`);
}

export async function forceDeleteLabel(id: number): Promise<void> {
  await apiClient.delete(`/api/v1/labels/${id}/force`);
}
