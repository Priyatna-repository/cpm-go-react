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

export async function listClientUsers(): Promise<ClientUser[]> {
  const { data } = await apiClient.get<ClientUser[]>('/api/v1/lookups/client-users');
  return data;
}