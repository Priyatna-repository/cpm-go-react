import { apiClient } from './client';

export interface Country {
  id: number;
  name: string;
}

export interface Currency {
  id: number;
  name: string;
  code: string;
  symbol: string;
  decimals: number;
}

export interface ClientUser {
    id: number;
    name: string;
}

export async function listCountries(): Promise<Country[]> {
  const { data } = await apiClient.get<Country[]>('/api/v1/lookups/countries');
  return data;
}

export async function listCurrencies(): Promise<Currency[]> {
  const { data } = await apiClient.get<Currency[]>('/api/v1/lookups/currencies');
  return data;
}

export async function listClientUsers(excludeCompanyId?: number): Promise<ClientUser[]> {
  const { data } = await apiClient.get<ClientUser[]>('/api/v1/lookups/client-users', {
    params: { company_id: excludeCompanyId },
  });
  return data;
}

export interface InternalUser {
  id: number;
  name: string;
}

export async function listInternalUsers(): Promise<InternalUser[]> {
  const { data } = await apiClient.get<InternalUser[]>('/api/v1/lookups/internal-users');
  return data;
}
