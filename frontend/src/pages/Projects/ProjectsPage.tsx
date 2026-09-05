import { useCallback, useEffect, useState, type FormEvent } from 'react';
import {
  ActionIcon, Alert, Badge, Button, Group, Modal, MultiSelect, NumberInput, Pagination,
  SegmentedControl, Select, Stack, Table, Text, TextInput, Textarea, Title, Tooltip,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import {
  IconAlertTriangle, IconArchive, IconEdit, IconPlus, IconRestore, IconTrash, IconUsers,
} from '@tabler/icons-react';
import * as api from '../../api/projects';
import type { Project, ProjectInput } from '../../api/projects';
import * as clientCompaniesApi from '../../api/clientCompanies';
import type { ClientCompany } from '../../api/clientCompanies';
import * as lookupsApi from '../../api/lookups';
import type { ClientUser, InternalUser } from '../../api/lookups';
import * as labelsApi from '../../api/labels';
import type { Label } from '../../api/labels';
import { useAuth } from '../../auth/AuthContext';
import { errorMessage } from '../../utils/errors';

const emptyForm: ProjectInput = { name: '', status_label_ids: [] };

export function ProjectsPage() {
  const { can } = useAuth();

  const [projects, setProjects] = useState<Project[]>([]);
  const [page, setPage] = useState(1);
  const [lastPage, setLastPage] = useState(1);
  const [searchInput, setSearchInput] = useState('');
  const [search] = useDebouncedValue(searchInput, 400);
  const [showArchived, setShowArchived] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [companies, setCompanies] = useState<ClientCompany[]>([]);
  const [clientUsers, setClientUsers] = useState<ClientUser[]>([]);
  const [internalUsers, setInternalUsers] = useState<InternalUser[]>([]);
  const [contractTypes, setContractTypes] = useState<Label[]>([]);
  const [statusOptions, setStatusOptions] = useState<Label[]>([]);

  const [modalOpen, setModalOpen] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [clientMode, setClientMode] = useState<'company' | 'individual'>('company');
  const [form, setForm] = useState<ProjectInput>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const [accessModalOpen, setAccessModalOpen] = useState(false);
  const [accessProjectId, setAccessProjectId] = useState<number | null>(null);
  const [accessUserIds, setAccessUserIds] = useState<string[]>([]);
  const [accessSaving, setAccessSaving] = useState(false);

  const loadProjects = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await api.listProjects({ page, search, archived: showArchived });
      setProjects(result.data);
      setLastPage(result.meta.last_page);
    } catch (err) {
      setError(errorMessage(err, 'Failed to load projects.'));
    } finally {
      setLoading(false);
    }
  }, [page, search, showArchived]);

  useEffect(() => {
    loadProjects();
  }, [loadProjects]);

  useEffect(() => {
    setPage(1);
  }, [search, showArchived]);

  useEffect(() => {
    (async () => {
      const results = await Promise.allSettled([
        clientCompaniesApi.listClientCompanies({ page: 1 }),
        lookupsApi.listClientUsers(),
        lookupsApi.listInternalUsers(),
        labelsApi.listLabels({ type: 'kontrak_label' }),
        labelsApi.listLabels({ type: 'pt_status' }),
      ]);
      if (results[0].status === 'fulfilled') setCompanies(results[0].value.data);
      if (results[1].status === 'fulfilled') setClientUsers(results[1].value);
      if (results[2].status === 'fulfilled') setInternalUsers(results[2].value);
      if (results[3].status === 'fulfilled') setContractTypes(results[3].value.data);
      if (results[4].status === 'fulfilled') setStatusOptions(results[4].value.data);
    })();
  }, []);

  function openCreate() {
    setEditingId(null);
    setClientMode('company');
    setForm(emptyForm);
    setFormError(null);
    setModalOpen(true);
  }

  function openEdit(project: Project) {
    setEditingId(project.id);
    setClientMode(project.client_user ? 'individual' : 'company');
    setForm({
      name: project.name,
      description: project.description,
      client_company_id: project.client_company?.id,
      client_user_id: project.client_user?.id,
      start_date: project.start_date,
      end_date: project.end_date,
      budget_estimate: project.budget_estimate,
      type_label_id: project.type_label?.id,
      status_label_ids: project.status_labels.map((l) => l.id),
    });
    setFormError(null);
    setModalOpen(true);
  }

  async function handleSave(e: FormEvent) {
    e.preventDefault();
    setSaving(true);
    setFormError(null);
    try {
      const payload: ProjectInput = {
        ...form,
        client_company_id: clientMode === 'company' ? form.client_company_id : undefined,
        client_user_id: clientMode === 'individual' ? form.client_user_id : undefined,
      };
      if (editingId) {
        await api.updateProject(editingId, payload);
      } else {
        await api.createProject(payload);
      }
      setModalOpen(false);
      await loadProjects();
    } catch (err) {
      setFormError(errorMessage(err, 'Failed to save.'));
    } finally {
      setSaving(false);
    }
  }

  async function handleArchive(id: number) {
    try {
      await api.archiveProject(id);
      await loadProjects();
    } catch (err) {
      setError(errorMessage(err, 'Failed to archive.'));
    }
  }

  async function handleRestore(id: number) {
    try {
      await api.restoreProject(id);
      await loadProjects();
    } catch (err) {
      setError(errorMessage(err, 'Failed to restore.'));
    }
  }

  async function handleForceDelete(id: number) {
    if (!window.confirm('Permanently delete this project? This cannot be undone.')) return;
    try {
      await api.forceDeleteProject(id);
      await loadProjects();
    } catch (err) {
      setError(errorMessage(err, 'Failed to delete.'));
    }
  }

  function openAccess(project: Project) {
    setAccessProjectId(project.id);
    setAccessUserIds(project.users.map((u) => String(u.id)));
    setAccessModalOpen(true);
  }

  async function handleSaveAccess() {
    if (accessProjectId === null) return;
    setAccessSaving(true);
    try {
      await api.updateProjectAccess(accessProjectId, accessUserIds.map(Number));
      setAccessModalOpen(false);
      await loadProjects();
    } catch (err) {
      setError(errorMessage(err, 'Failed to update access.'));
    } finally {
      setAccessSaving(false);
    }
  }

  const selectedCompany = form.client_company_id
   ? companies.find((c) => c.id === form.client_company_id)
   : undefined;
  const canManageAccess = can('project.manage_access');
  const clientLocked = editingId !== null && !canManageAccess;

  return (
    <Stack p="xl">
      <Group justify="space-between">
        <Title order={2}>Projects</Title>
        {can('project.create') && (
          <Button leftSection={<IconPlus size={16} />} onClick={openCreate}>
            Create
          </Button>
        )}
      </Group>

      <Group>
        <TextInput
          placeholder="Search by name or code"
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

      <Table.ScrollContainer minWidth={900}>
        <Table>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Code</Table.Th>
              <Table.Th>Name</Table.Th>
              <Table.Th>Client</Table.Th>
              <Table.Th>Dates</Table.Th>
              <Table.Th>Status</Table.Th>
              <Table.Th>Actions</Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {projects.map((project) => (
              <Table.Tr key={project.id}>
                <Table.Td>
                  <Text size="sm" ff="monospace">{project.code}</Text>
                </Table.Td>
                <Table.Td>{project.name}</Table.Td>
                <Table.Td>{project.client_company?.name ?? project.client_user?.name ?? '—'}</Table.Td>
                <Table.Td>
                  <Text size="sm">{project.start_date ?? '—'} → {project.end_date ?? '—'}</Text>
                </Table.Td>
                <Table.Td>
                  <Group gap={4}>
                    {project.status_labels.map((l) => (
                      <Badge key={l.id} style={l.color ? { backgroundColor: l.color } : undefined}>
                        {l.name}
                      </Badge>
                    ))}
                  </Group>
                </Table.Td>
                <Table.Td>
                  <Group gap="xs">
                    {can('project.edit') && !project.archived_at && (
                      <Tooltip label="Edit">
                        <ActionIcon variant="subtle" onClick={() => openEdit(project)}>
                          <IconEdit size={16} />
                        </ActionIcon>
                      </Tooltip>
                    )}
                    {can('project.manage_access') && !project.archived_at && (
                      <Tooltip label="Manage access">
                        <ActionIcon variant="subtle" onClick={() => openAccess(project)}>
                          <IconUsers size={16} />
                        </ActionIcon>
                      </Tooltip>
                    )}
                    {can('project.archive') && !project.archived_at && (
                      <Tooltip label="Archive">
                        <ActionIcon variant="subtle" color="orange" onClick={() => handleArchive(project.id)}>
                          <IconArchive size={16} />
                        </ActionIcon>
                      </Tooltip>
                    )}
                    {can('project.restore') && project.archived_at && (
                      <Tooltip label="Restore">
                        <ActionIcon variant="subtle" color="green" onClick={() => handleRestore(project.id)}>
                          <IconRestore size={16} />
                        </ActionIcon>
                      </Tooltip>
                    )}
                    {can('project.delete') && project.archived_at && (
                      <Tooltip label="Delete permanently">
                        <ActionIcon variant="subtle" color="red" onClick={() => handleForceDelete(project.id)}>
                          <IconTrash size={16} />
                        </ActionIcon>
                      </Tooltip>
                    )}
                  </Group>
                </Table.Td>
              </Table.Tr>
            ))}
          </Table.Tbody>
        </Table>
      </Table.ScrollContainer>

      {!loading && projects.length === 0 && <Text c="dimmed">No projects found.</Text>}

      <Group justify="center">
        <Pagination value={page} onChange={setPage} total={lastPage} />
      </Group>

      <Modal opened={modalOpen} onClose={() => setModalOpen(false)} title={editingId ? 'Edit project' : 'Create project'} size="lg">
        <form onSubmit={handleSave}>
          <Stack>
            {formError && (
              <Alert color="red" icon={<IconAlertTriangle size={18} />}>
                {formError}
              </Alert>
            )}
            <TextInput
              label="Name"
              required
              value={form.name}
              onChange={(e) => setForm((f) => ({ ...f, name: e.currentTarget.value }))}
            />
            <Textarea
              label="Description"
              value={form.description ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, description: e.currentTarget.value }))}
              autosize
              minRows={2}
            />

              <div>
              <Text size="sm" fw={500} mb={2}>Client</Text>
              <Text size="xs" c="dimmed" mb={6}>
                Whoever is linked here automatically gets access to view this project —
                either everyone in a company, or just one person.
              </Text>
              <SegmentedControl
                fullWidth
                disabled={clientLocked}
                value={clientMode}
                onChange={(v) => setClientMode(v as 'company' | 'individual')}
                data={[
                  { label: 'Client company (all its members get access)', value: 'company' },
                  { label: 'Individual client (only this person)', value: 'individual' },
                ]}
              />
             {clientLocked && (
                <Text size="xs" c="orange" mb={6}>
                  You don't have permission to change this project's client — contact someone with
                  "manage access" rights if it needs to change.
                </Text>
              )}
            </div>
            {clientMode === 'company' ? (
              <>
                <Select
                  label="Client company"
                  searchable
                  disabled={clientLocked}
                  data={companies.map((c) => ({ value: String(c.id), label: c.name }))}
                  value={form.client_company_id ? String(form.client_company_id) : null}
                  onChange={(v) => setForm((f) => ({ ...f, client_company_id: v ? Number(v) : undefined }))}
                />
                {selectedCompany && (
                  <Alert color="blue" variant="light">
                    {selectedCompany.clients.length > 0 ? (
                      <>
                        Will give project access to {selectedCompany.clients.length}{' '}
                        member{selectedCompany.clients.length > 1 ? 's' : ''} of{' '}
                        <strong>{selectedCompany.name}</strong>:{' '}
                        {selectedCompany.clients.map((c) => c.name).join(', ')}
                      </>
                    ) : (
                      <>
                        <strong>{selectedCompany.name}</strong> has no members assigned yet — no
                        one will get project access via this company until you add members on the
                        Client Companies page.
                      </>
                    )}
                  </Alert>
                )}
              </>
            ) : (
              <>
                <Select
                  label="Individual client"
                  searchable
                  disabled={clientLocked}
                  data={clientUsers.map((u) => ({ value: String(u.id), label: u.name }))}
                  value={form.client_user_id ? String(form.client_user_id) : null}
                  onChange={(v) => setForm((f) => ({ ...f, client_user_id: v ? Number(v) : undefined }))}
                />
                {form.client_user_id && (
                  <Alert color="blue" variant="light">
                    Only this one person gets project access via the client relationship — grant
                    more internal users access separately with "Manage access" after creating.
                  </Alert>
                )}
              </>
            )}

            <Group grow>
              <TextInput
                type="date"
                label="Start date"
                value={form.start_date ?? ''}
                onChange={(e) => setForm((f) => ({ ...f, start_date: e.currentTarget.value }))}
              />
              <TextInput
                type="date"
                label="End date"
                value={form.end_date ?? ''}
                onChange={(e) => setForm((f) => ({ ...f, end_date: e.currentTarget.value }))}
              />
            </Group>

            <NumberInput
              label="Budget estimate"
              value={form.budget_estimate}
              onChange={(v) => setForm((f) => ({ ...f, budget_estimate: typeof v === 'number' ? v : undefined }))}
              min={0}
              decimalScale={2}
            />

            <Select
              label="Contract type"
              searchable
              clearable
              data={contractTypes.map((l) => ({ value: String(l.id), label: l.name }))}
              value={form.type_label_id ? String(form.type_label_id) : null}
              onChange={(v) => setForm((f) => ({ ...f, type_label_id: v ? Number(v) : undefined }))}
            />

            <MultiSelect
              label="Status"
              searchable
              data={statusOptions.map((l) => ({ value: String(l.id), label: l.name }))}
              value={form.status_label_ids.map(String)}
              onChange={(values) => setForm((f) => ({ ...f, status_label_ids: values.map(Number) }))}
            />

            <Group justify="flex-end" mt="md">
              <Button variant="default" onClick={() => setModalOpen(false)}>Cancel</Button>
              <Button type="submit" loading={saving}>Save</Button>
            </Group>
          </Stack>
        </form>
      </Modal>

      <Modal opened={accessModalOpen} onClose={() => setAccessModalOpen(false)} title="Manage project access">
        <Stack>
          <Text size="sm" c="dimmed">
            Grant specific internal users (manager/team member) access to this project. Client-company
            members and the project's individual client already have access automatically.
          </Text>
          <MultiSelect
            label="Users with access"
            searchable
            data={internalUsers.map((u) => ({ value: String(u.id), label: u.name }))}
            value={accessUserIds}
            onChange={setAccessUserIds}
          />
          <Group justify="flex-end" mt="md">
            <Button variant="default" onClick={() => setAccessModalOpen(false)}>Cancel</Button>
            <Button onClick={handleSaveAccess} loading={accessSaving}>Save</Button>
          </Group>
        </Stack>
      </Modal>
    </Stack>
  );
}
