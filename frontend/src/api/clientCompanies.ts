import { apiClient } from './client';
import { appendIfPresent } from '../utils/formData';

export interface ClientCompanyUser {
  id: number;
  name: string;
}

export interface ClientCompany {
  id: number;
  name: string;
  logo?: string;
  address?: string;
  postal_code?: string;
  city?: string;
  country_id?: number;
  currency_id?: number;
  email?: string;
  phone?: string;
  web?: string;
  clients: ClientCompanyUser[];
  archived_at?: string;
}

export interface ClientCompanyListResponse {
  data: ClientCompany[];
  meta: { current_page: number; last_page: number; total: number };
}

export interface ListParams {
  page?: number;
  search?: string;
  archived?: boolean;
}

export async function listClientCompanies(params: ListParams = {}): Promise<ClientCompanyListResponse> {
  const { data } = await apiClient.get<ClientCompanyListResponse>('/api/v1/client-companies', {
    params: {
      page: params.page,
      search: params.search || undefined,
      archived: params.archived ? '1' : undefined,
    },
  });
  return data;
}

export interface ClientCompanyInput {
  name: string;
  address?: string;
  postal_code?: string;
  city?: string;
  country_id?: number;
  currency_id?: number;
  email?: string;
  phone?: string;
  web?: string;
  client_ids: number[];
  logo?: File | null;
}

function toFormData(input: ClientCompanyInput): FormData {
  const form = new FormData();
  form.append('name', input.name);
  appendIfPresent(form, 'address', input.address);
  appendIfPresent(form, 'postal_code', input.postal_code);
  appendIfPresent(form, 'city', input.city);
  appendIfPresent(form, 'country_id', input.country_id);
  appendIfPresent(form, 'currency_id', input.currency_id);
  appendIfPresent(form, 'email', input.email);
  appendIfPresent(form, 'phone', input.phone);
  appendIfPresent(form, 'web', input.web);
  input.client_ids.forEach((id) => form.append('client_ids', String(id)));
  if (input.logo) form.append('logo', input.logo);
  return form;
}

export async function createClientCompany(input: ClientCompanyInput): Promise<ClientCompany> {
  const { data } = await apiClient.post<ClientCompany>('/api/v1/client-companies', toFormData(input), {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return data;
}

export async function updateClientCompany(id: number, input: ClientCompanyInput): Promise<ClientCompany> {
  const { data } = await apiClient.put<ClientCompany>(`/api/v1/client-companies/${id}`, toFormData(input), {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return data;
}

export async function archiveClientCompany(id: number): Promise<void> {
  await apiClient.delete(`/api/v1/client-companies/${id}`);
}

export async function restoreClientCompany(id: number): Promise<void> {
  await apiClient.post(`/api/v1/client-companies/${id}/restore`);
}

export async function forceDeleteClientCompany(id: number): Promise<void> {
  await apiClient.delete(`/api/v1/client-companies/${id}/force`);
}
