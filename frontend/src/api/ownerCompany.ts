import { apiClient } from './client';
import { appendIfPresent } from '../utils/formData';

export interface OwnerCompany {
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
}

export async function getOwnerCompany(): Promise<OwnerCompany> {
  const { data } = await apiClient.get<OwnerCompany>('/api/v1/owner-company');
  return data;
}

export interface UpdateOwnerCompanyInput {
  name: string;
  address?: string;
  postal_code?: string;
  city?: string;
  country_id?: number;
  currency_id?: number;
  email?: string;
  phone?: string;
  web?: string;
  logo?: File | null;
}

export async function updateOwnerCompany(input: UpdateOwnerCompanyInput): Promise<OwnerCompany> {
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
  if (input.logo) form.append('logo', input.logo);

  const { data } = await apiClient.put<OwnerCompany>('/api/v1/owner-company', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return data;
}
