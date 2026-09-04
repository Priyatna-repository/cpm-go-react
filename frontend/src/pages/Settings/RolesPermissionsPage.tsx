import { useEffect, useMemo, useState } from 'react';
import axios from 'axios';
import { errorMessage } from '../../utils/errors';
import {
  Alert,
  Badge,
  Button,
  Center,
  Checkbox,
  Fieldset,
  Group,
  Loader,
  Stack,
  Text,
  Title,
} from '@mantine/core';
import { IconAlertTriangle } from '@tabler/icons-react';
import * as permissionsApi from '../../api/permissions';
import type { Permission, RoleWithPermissions } from '../../api/permissions';

export function RolesPermissionsPage() {
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [roles, setRoles] = useState<RoleWithPermissions[]>([]);
  const [draft, setDraft] = useState<Record<number, Set<string>>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [savingRoleId, setSavingRoleId] = useState<number | null>(null);

  useEffect(() => {
    (async () => {
      try {
        const [perms, rolesList] = await Promise.all([
          permissionsApi.listPermissions(),
          permissionsApi.listRoles(),
        ]);
        setPermissions(perms);
        setRoles(rolesList);
        setDraft(Object.fromEntries(rolesList.map((r) => [r.id, new Set(r.permissions)])));
      } catch {
        setError('Failed to load roles and permissions.');
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const grouped = useMemo(() => {
    const map = new Map<string, Permission[]>();
    for (const p of permissions) {
      const list = map.get(p.category) ?? [];
      list.push(p);
      map.set(p.category, list);
    }
    return Array.from(map.entries());
  }, [permissions]);

  function toggle(roleId: number, permissionName: string) {
    setDraft((prev) => {
      const next = new Set(prev[roleId]);
      if (next.has(permissionName)) {
        next.delete(permissionName);
      } else {
        next.add(permissionName);
      }
      return { ...prev, [roleId]: next };
    });
  }

  async function save(roleId: number) {
    const role = roles.find((r) => r.id === roleId);
    if (!role) return;

    setSavingRoleId(roleId);
    setError(null);
    try {
      const updated = await permissionsApi.updateRolePermissions(
        roleId,
        role.version,
        Array.from(draft[roleId] ?? []),
      );
      setRoles((prev) => prev.map((r) => (r.id === roleId ? updated : r)));
      setDraft((prev) => ({ ...prev, [roleId]: new Set(updated.permissions) }));
      } catch (err) {
      if (axios.isAxiosError(err) && err.response?.status === 409) {
        setError(`"${role.name}" was changed by someone else — reload the page and try again.`);
      } else {
        setError(errorMessage(err, 'Failed to save changes.'));
      }
    } finally {
      setSavingRoleId(null);
    }
  }

  if (loading) {
    return (
      <Center p="xl">
        <Loader />
      </Center>
    );
  }

  return (
    <Stack p="xl">
      <Title order={2}>Roles & Permissions</Title>
      <Text c="dimmed" size="sm">
        Admin always has full access and isn&apos;t shown here. Toggle which permissions each
        role has, then save that role.
      </Text>

      {error && (
        <Alert color="red" icon={<IconAlertTriangle size={18} />}>
          {error}
        </Alert>
      )}

      {roles.map((role) => (
        <Fieldset key={role.id} legend={<Badge tt="capitalize">{role.name}</Badge>}>
          <Stack gap="md">
            {grouped.map(([category, perms]) => (
              <div key={category}>
                <Text fw={600} size="sm" mb={4}>
                  {category}
                </Text>
                <Group gap="lg">
                  {perms.map((perm) => (
                    <Checkbox
                      key={perm.name}
                      label={perm.name}
                      checked={draft[role.id]?.has(perm.name) ?? false}
                      onChange={() => toggle(role.id, perm.name)}
                    />
                  ))}
                </Group>
              </div>
            ))}
            <Group justify="flex-end">
              <Button loading={savingRoleId === role.id} onClick={() => save(role.id)}>
                Save {role.name}
              </Button>
            </Group>
          </Stack>
        </Fieldset>
      ))}
    </Stack>
  );
}
