import { apiClient } from './client';

export interface Permission {
  id: number;
  name: string;
  category: string;
  description?: string;
}

export interface RoleWithPermissions {
  id: number;
  name: string;
  version: number;
  permissions: string[];
}

export async function listPermissions(): Promise<Permission[]> {
  const { data } = await apiClient.get<Permission[]>('/api/v1/permissions');
  return data;
}

export async function listRoles(): Promise<RoleWithPermissions[]> {
  const { data } = await apiClient.get<RoleWithPermissions[]>('/api/v1/roles');
  return data;
}

export async function updateRolePermissions(
  roleId: number,
  version: number,
  permissions: string[],
): Promise<RoleWithPermissions> {
  const { data } = await apiClient.put<RoleWithPermissions>(
    `/api/v1/roles/${roleId}/permissions`,
    { version, permissions },
  );
  return data;
}
