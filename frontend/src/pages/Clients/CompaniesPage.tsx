import { useCallback, useEffect, useState, type FormEvent } from 'react';
import {
  ActionIcon, Alert, Avatar, Button, Group, Modal, MultiSelect, Pagination,
  Select, Stack, Table, Text, TextInput, Title, Tooltip,
} from '@mantine/core';
import {
  IconAlertTriangle, IconArchive, IconEdit, IconPlus, IconRestore, IconTrash,
} from '@tabler/icons-react';
import * as api from '../../api/clientCompanies';
import type { ClientCompany, ClientCompanyInput } from '../../api/clientCompanies';
import * as lookupsApi from '../../api/lookups';
import type { Country, Currency, ClientUser } from '../../api/lookups';
import { useAuth } from '../../auth/AuthContext';
import { errorMessage } from '../../utils/errors';

const emptyForm: ClientCompanyInput = { name: '', client_ids: [], logo: null };

export function CompaniesPage() {
  const { can } = useAuth();

  const [companies, setCompanies] = useState<ClientCompany[]>([]);
  const [page, setPage] = useState(1);
  const [lastPage, setLastPage] = useState(1);
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [showArchived, setShowArchived] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [countries, setCountries] = useState<Country[]>([]);
  const [currencies, setCurrencies] = useState<Currency[]>([]);
  const [clientUsers, setClientUsers] = useState<ClientUser[]>([]);

  const [modalOpen, setModalOpen] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [form, setForm] = useState<ClientCompanyInput>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

    useEffect(() => {
    const timeout = setTimeout(() => {
      setPage(1);
      setSearch(searchInput);
    }, 400);
    return () => clearTimeout(timeout);
  }, [searchInput]);

  const loadCompanies = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await api.listClientCompanies({ page, search, archived: showArchived });
      setCompanies(result.data);
      setLastPage(result.meta.last_page);
    } catch (err) {
      setError(errorMessage(err, 'Failed to load client companies.'));
    } finally {
      setLoading(false);
    }
  }, [page, search, showArchived]);

  useEffect(() => {
    loadCompanies();
  }, [loadCompanies]);

  useEffect(() => {
    (async () => {
      const [countryList, currencyList, userList] = await Promise.all([
        lookupsApi.listCountries(),
        lookupsApi.listCurrencies(),
        lookupsApi.listClientUsers(),
      ]);
      setCountries(countryList);
      setCurrencies(currencyList);
      setClientUsers(userList);
    })();
  }, []);

  function openCreate() {
    setEditingId(null);
    setForm(emptyForm);
    setFormError(null);
    setModalOpen(true);
  }

  function openEdit(company: ClientCompany) {
    setEditingId(company.id);
    setForm({
      name: company.name,
      address: company.address,
      postal_code: company.postal_code,
      city: company.city,
      country_id: company.country_id,
      currency_id: company.currency_id,
      email: company.email,
      phone: company.phone,
      web: company.web,
      client_ids: company.clients.map((c) => c.id),
      logo: null,
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
        await api.updateClientCompany(editingId, form);
      } else {
        await api.createClientCompany(form);
      }
      setModalOpen(false);
      await loadCompanies();
    } catch (err) {
      setFormError(errorMessage(err, 'Failed to save.'));
    } finally {
      setSaving(false);
    }
  }

  async function handleArchive(id: number) {
    try {
      await api.archiveClientCompany(id);
      await loadCompanies();
    } catch (err) {
      setError(errorMessage(err, 'Failed to archive.'));
    }
  }

  async function handleRestore(id: number) {
    try {
      await api.restoreClientCompany(id);
      await loadCompanies();
    } catch (err) {
      setError(errorMessage(err, 'Failed to restore.'));
    }
  }

  async function handleForceDelete(id: number) {
    if (!window.confirm('Permanently delete this client company? This cannot be undone.')) return;
    try {
      await api.forceDeleteClientCompany(id);
      await loadCompanies();
    } catch (err) {
      setError(errorMessage(err, 'Failed to delete.'));
    }
  }

  return (
    <Stack p="xl">
      <Group justify="space-between">
        <Title order={2}>Client Companies</Title>
        {can('client_company.create') && (
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
          onClick={() => { setPage(1); setShowArchived((v) => !v); }}
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
              <Table.Th>Logo</Table.Th>
              <Table.Th>Name</Table.Th>
              <Table.Th>Email</Table.Th>
              <Table.Th>Currency</Table.Th>
              <Table.Th>Clients</Table.Th>
              <Table.Th>Actions</Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {companies.map((company) => (
              <Table.Tr key={company.id}>
                <Table.Td>
                  <Avatar
                    src={company.logo ? `${import.meta.env.VITE_API_BASE_URL}${company.logo}` : undefined}
                    radius="xl"
                  >
                    {company.name.slice(0, 1)}
                  </Avatar>
                </Table.Td>
                <Table.Td>{company.name}</Table.Td>
                <Table.Td>{company.email ?? '—'}</Table.Td>
                <Table.Td>{currencies.find((c) => c.id === company.currency_id)?.code ?? '—'}</Table.Td>
                <Table.Td>
                  <Text size="sm">{company.clients.length}</Text>
                </Table.Td>
                <Table.Td>
                  <Group gap="xs">
                    {can('client_company.edit') && !company.archived_at && (
                      <Tooltip label="Edit">
                        <ActionIcon variant="subtle" onClick={() => openEdit(company)}>
                          <IconEdit size={16} />
                        </ActionIcon>
                      </Tooltip>
                    )}
                    {can('client_company.archive') && !company.archived_at && (
                      <Tooltip label="Archive">
                        <ActionIcon variant="subtle" color="orange" onClick={() => handleArchive(company.id)}>
                          <IconArchive size={16} />
                        </ActionIcon>
                      </Tooltip>
                    )}
                    {can('client_company.restore') && company.archived_at && (
                      <Tooltip label="Restore">
                        <ActionIcon variant="subtle" color="green" onClick={() => handleRestore(company.id)}>
                          <IconRestore size={16} />
                        </ActionIcon>
                      </Tooltip>
                    )}
                    {can('client_company.delete') && company.archived_at && (
                      <Tooltip label="Delete permanently">
                        <ActionIcon variant="subtle" color="red" onClick={() => handleForceDelete(company.id)}>
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

      {!loading && companies.length === 0 && <Text c="dimmed">No client companies found.</Text>}

      <Group justify="center">
        <Pagination value={page} onChange={setPage} total={lastPage} />
      </Group>

      <Modal opened={modalOpen} onClose={() => setModalOpen(false)} title={editingId ? 'Edit client company' : 'Create client company'} size="lg">
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
            <TextInput
              label="Email"
              value={form.email ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, email: e.currentTarget.value }))}
            />
            <Group grow>
              <TextInput
                label="Phone"
                value={form.phone ?? ''}
                onChange={(e) => setForm((f) => ({ ...f, phone: e.currentTarget.value }))}
              />
              <TextInput
                label="Website"
                value={form.web ?? ''}
                onChange={(e) => setForm((f) => ({ ...f, web: e.currentTarget.value }))}
              />
            </Group>
            <TextInput
              label="Address"
              value={form.address ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, address: e.currentTarget.value }))}
            />
            <Group grow>
              <TextInput
                label="City"
                value={form.city ?? ''}
                onChange={(e) => setForm((f) => ({ ...f, city: e.currentTarget.value }))}
              />
              <TextInput
                label="Postal code"
                value={form.postal_code ?? ''}
                onChange={(e) => setForm((f) => ({ ...f, postal_code: e.currentTarget.value }))}
              />
            </Group>
            <Group grow>
              <Select
                label="Country"
                searchable
                data={countries.map((c) => ({ value: String(c.id), label: c.name }))}
                value={form.country_id ? String(form.country_id) : null}
                onChange={(v) => setForm((f) => ({ ...f, country_id: v ? Number(v) : undefined }))}
              />
              <Select
                label="Currency"
                searchable
                data={currencies.map((c) => ({ value: String(c.id), label: `${c.name} (${c.symbol})` }))}
                value={form.currency_id ? String(form.currency_id) : null}
                onChange={(v) => setForm((f) => ({ ...f, currency_id: v ? Number(v) : undefined }))}
              />
            </Group>
            <MultiSelect
              label="Assigned clients"
              placeholder="Select client users"
              data={clientUsers.map((u) => ({ value: String(u.id), label: u.name }))}
              value={form.client_ids.map(String)}
              onChange={(values) => setForm((f) => ({ ...f, client_ids: values.map(Number) }))}
              searchable
            />
            <div>
              <Text size="sm" fw={500} mb={4}>Logo</Text>
              <input
                type="file"
                accept="image/png,image/jpeg,image/gif"
                onChange={(e) => setForm((f) => ({ ...f, logo: e.target.files?.[0] ?? null }))}
              />
            </div>
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
