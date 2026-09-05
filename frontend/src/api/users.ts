import { apiClient } from './client';

export interface AppUser {
  id: number;
  name: string;
  email: string;
  role: string;
  avatar?: string;
  job_title?: string;
  phone?: string;
  address?: string;
  archived_at?: string;
}

export interface UserListResponse {
  data: AppUser[];
  meta: { current_page: number; last_page: number; total: number };
}

export interface ListUsersParams {
  page?: number;
  search?: string;
  archived?: boolean;
}

export interface UserInput {
  name: string;
  email: string;
  password?: string;
  role: string;
  job_title?: string;
  phone?: string;
  address?: string;
  avatar?: File | null;
}

function toFormData(input: UserInput): FormData {
  const form = new FormData();
  form.append('name', input.name);
  form.append('email', input.email);
  form.append('role', input.role);
  if (input.password) form.append('password', input.password);
  if (input.job_title) form.append('job_title', input.job_title);
  if (input.phone) form.append('phone', input.phone);
  if (input.address) form.append('address', input.address);
  if (input.avatar) form.append('avatar', input.avatar);
  return form;
}

export function createUsersApi(basePath: string) {
  return {
    async list(params: ListUsersParams = {}): Promise<UserListResponse> {
      const { data } = await apiClient.get<UserListResponse>(basePath, {
        params: {
          page: params.page,
          search: params.search || undefined,
          archived: params.archived ? '1' : undefined,
        },
      });
      return data;
    },
    async create(input: UserInput): Promise<AppUser> {
      const { data } = await apiClient.post<AppUser>(basePath, toFormData(input), {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      return data;
    },
    async update(id: number, input: UserInput): Promise<AppUser> {
      const { data } = await apiClient.put<AppUser>(`${basePath}/${id}`, toFormData(input), {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      return data;
    },
    async archive(id: number): Promise<void> {
      await apiClient.delete(`${basePath}/${id}`);
    },
    async restore(id: number): Promise<void> {
      await apiClient.post(`${basePath}/${id}/restore`);
    },
    async forceDelete(id: number): Promise<void> {
      await apiClient.delete(`${basePath}/${id}/force`);
    },
  };
}

export const internalUsersApi = createUsersApi('/api/v1/users');
export const clientUserAccountsApi = createUsersApi('/api/v1/client-user-accounts');
