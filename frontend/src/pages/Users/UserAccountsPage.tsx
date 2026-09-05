import { useCallback, useEffect, useState, type FormEvent } from 'react';
import {
  ActionIcon, Alert, Avatar, Button, Group, Modal, Pagination, PasswordInput,
  Select, Stack, Table, Text, TextInput, Title, Tooltip,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import {
  IconAlertTriangle, IconArchive, IconEdit, IconPlus, IconRestore, IconTrash,
} from '@tabler/icons-react';
import { createUsersApi, type AppUser, type UserInput } from '../../api/users';
import { LogoUpload } from '../../components/LogoUpload';
import { useAuth } from '../../auth/AuthContext';
import { errorMessage } from '../../utils/errors';

interface UserAccountsPageProps {
  api: ReturnType<typeof createUsersApi>;
  title: string;
  permissionPrefix: string;
  roleOptions?: { value: string; label: string }[];
  fixedRole?: string;
}

function emptyForm(fixedRole?: string): UserInput {
  return { name: '', email: '', role: fixedRole ?? '', avatar: null };
}

export function UserAccountsPage({ api, title, permissionPrefix, roleOptions, fixedRole }: UserAccountsPageProps) {
  const { user: currentUser, can } = useAuth();
  const canView = can(`${permissionPrefix}.view`);
  const canCreate = can(`${permissionPrefix}.create`);
  const canEdit = can(`${permissionPrefix}.edit`);
  const canArchive = can(`${permissionPrefix}.archive`);
  const canRestore = can(`${permissionPrefix}.restore`);
  const canDelete = can(`${permissionPrefix}.delete`);

  const [users, setUsers] = useState<AppUser[]>([]);
  const [page, setPage] = useState(1);
  const [lastPage, setLastPage] = useState(1);
  const [searchInput, setSearchInput] = useState('');
  const [search] = useDebouncedValue(searchInput, 400);
  const [showArchived, setShowArchived] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [modalOpen, setModalOpen] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [form, setForm] = useState<UserInput>(emptyForm(fixedRole));
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const loadUsers = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await api.list({ page, search, archived: showArchived });
      setUsers(result.data);
      setLastPage(result.meta.last_page);
    } catch (err) {
      setError(errorMessage(err, 'Failed to load users.'));
    } finally {
      setLoading(false);
    }
  }, [api, page, search, showArchived]);

  useEffect(() => {
    loadUsers();
  }, [loadUsers]);

  useEffect(() => {
    setPage(1);
  }, [search, showArchived]);

  function openCreate() {
    setEditingId(null);
    setForm(emptyForm(fixedRole));
    setFormError(null);
    setModalOpen(true);
  }

  function openEdit(u: AppUser) {
    setEditingId(u.id);
    setForm({
      name: u.name, email: u.email, role: u.role, job_title: u.job_title,
      phone: u.phone, address: u.address, avatar: null,
    });
    setFormError(null);
    setModalOpen(true);
  }

  async function handleSave(e: FormEvent) {
    e.preventDefault();
    setSaving(true);
    setFormError(null);
    try {
      if (editingId) {
        await api.update(editingId, form);
      } else {
        await api.create(form);
      }
      setModalOpen(false);
      await loadUsers();
    } catch (err) {
      setFormError(errorMessage(err, 'Failed to save.'));
    } finally {
      setSaving(false);
    }
  }

  async function handleArchive(id: number) {
    try {
      await api.archive(id);
      await loadUsers();
    } catch (err) {
      setError(errorMessage(err, 'Failed to archive.'));
    }
  }

  async function handleRestore(id: number) {
    try {
      await api.restore(id);
      await loadUsers();
    } catch (err) {
      setError(errorMessage(err, 'Failed to restore.'));
    }
  }

  async function handleForceDelete(id: number) {
    if (!window.confirm('Permanently delete this user? This cannot be undone.')) return;
    try {
      await api.forceDelete(id);
      await loadUsers();
    } catch (err) {
      setError(errorMessage(err, 'Failed to delete.'));
    }
  }

  if (!canView) {
    return null;
  }

  const editingUser = editingId ? users.find((u) => u.id === editingId) : undefined;

  return (
    <Stack p="xl">
      <Group justify="space-between">
        <Title order={2}>{title}</Title>
        {canCreate && (
          <Button leftSection={<IconPlus size={16} />} onClick={openCreate}>
            Create
          </Button>
        )}
      </Group>

      <Group>
        <TextInput
          placeholder="Search by name or email"
          value={searchInput}
          onChange={(e) => setSearchInput(e.currentTarget.value)}
          w={280}
        />
        <Button
          variant={showArchived ? 'filled' : 'default'}
          onClick={() => setShowArchived((v) => !v)}
        >
          {showArchived ? 'Showing archived' : 'Show archived'}
        </Button>
      </Group>

      {error && (
        <Alert color="red" icon={<IconAlertTriangle size={18} />}>
          {error}
        </Alert>
      )}

      <Table.ScrollContainer minWidth={800}>
        <Table>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>User</Table.Th>
              <Table.Th>Email</Table.Th>
              {!fixedRole && <Table.Th>Role</Table.Th>}
              <Table.Th>Phone</Table.Th>
              <Table.Th>Actions</Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {users.map((u) => {
              const isSelf = currentUser?.id === u.id;
              return (
                <Table.Tr key={u.id}>
                  <Table.Td>
                    <Group gap="sm">
                      <Avatar
                        src={u.avatar ? `${import.meta.env.VITE_API_BASE_URL}${u.avatar}` : undefined}
                        radius="xl"
                        size="sm"
                      >
                        {u.name.slice(0, 1)}
                      </Avatar>
                      <Text size="sm">{u.name}</Text>
                    </Group>
                  </Table.Td>
                  <Table.Td>{u.email}</Table.Td>
                  {!fixedRole && <Table.Td>{u.role}</Table.Td>}
                  <Table.Td>{u.phone ?? '—'}</Table.Td>
                  <Table.Td>
                    <Group gap="xs">
                      {canEdit && !u.archived_at && (
                        <Tooltip label="Edit">
                          <ActionIcon variant="subtle" onClick={() => openEdit(u)}>
                            <IconEdit size={16} />
                          </ActionIcon>
                        </Tooltip>
                      )}
                      {canArchive && !u.archived_at && !isSelf && (
                        <Tooltip label="Archive">
                          <ActionIcon variant="subtle" color="orange" onClick={() => handleArchive(u.id)}>
                            <IconArchive size={16} />
                          </ActionIcon>
                        </Tooltip>
                      )}
                      {canRestore && u.archived_at && (
                        <Tooltip label="Restore">
                          <ActionIcon variant="subtle" color="green" onClick={() => handleRestore(u.id)}>
                            <IconRestore size={16} />
                          </ActionIcon>
                        </Tooltip>
                      )}
                      {canDelete && u.archived_at && !isSelf && (
                        <Tooltip label="Delete permanently">
                          <ActionIcon variant="subtle" color="red" onClick={() => handleForceDelete(u.id)}>
                            <IconTrash size={16} />
                          </ActionIcon>
                        </Tooltip>
                      )}
                    </Group>
                  </Table.Td>
                </Table.Tr>
              );
            })}
          </Table.Tbody>
        </Table>
      </Table.ScrollContainer>

      {!loading && users.length === 0 && <Text c="dimmed">No users found.</Text>}

      <Group justify="center">
        <Pagination value={page} onChange={setPage} total={lastPage} />
      </Group>

      <Modal opened={modalOpen} onClose={() => setModalOpen(false)} title={editingId ? 'Edit user' : 'Create user'} size="lg">
        <form onSubmit={handleSave}>
          <Stack>
            {formError && (
              <Alert color="red" icon={<IconAlertTriangle size={18} />}>
                {formError}
              </Alert>
            )}

            <LogoUpload
              currentUrl={editingUser?.avatar ? `${import.meta.env.VITE_API_BASE_URL}${editingUser.avatar}` : undefined}
              fallbackText={form.name || '?'}
              onChange={(file) => setForm((f) => ({ ...f, avatar: file }))}
            />

            <TextInput
              label="Name"
              required
              value={form.name}
              onChange={(e) => setForm((f) => ({ ...f, name: e.currentTarget.value }))}
            />
            <TextInput
              label="Email"
              required
              type="email"
              value={form.email}
              onChange={(e) => setForm((f) => ({ ...f, email: e.currentTarget.value }))}
            />
            <PasswordInput
              label="Password"
              required={!editingId}
              description={editingId ? 'Leave blank to keep the current password' : undefined}
              value={form.password ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, password: e.currentTarget.value }))}
            />
            {roleOptions && (
              <Select
                label="Role"
                required
                data={roleOptions}
                value={form.role || null}
                onChange={(v) => setForm((f) => ({ ...f, role: v ?? '' }))}
              />
            )}
            <TextInput
              label="Job title"
              value={form.job_title ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, job_title: e.currentTarget.value }))}
            />
            <Group grow>
              <TextInput
                label="Phone"
                value={form.phone ?? ''}
                onChange={(e) => setForm((f) => ({ ...f, phone: e.currentTarget.value }))}
              />
              <TextInput
                label="Address"
                value={form.address ?? ''}
                onChange={(e) => setForm((f) => ({ ...f, address: e.currentTarget.value }))}
              />
            </Group>

            <Group justify="flex-end" mt="md">
              <Button variant="default" onClick={() => setModalOpen(false)}>Cancel</Button>
              <Button type="submit" loading={saving}>Save</Button>
            </Group>
          </Stack>
        </form>
      </Modal>
    </Stack>
  );
}
