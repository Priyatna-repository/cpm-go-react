import { useCallback, useEffect, useState, type FormEvent } from 'react';
import {
  ActionIcon, Alert, Badge, Button, ColorInput, Group, Modal, Pagination,
  Select, Stack, Table, Text, TextInput, Title, Tooltip,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import {
  IconAlertTriangle, IconArchive, IconEdit, IconPlus, IconRestore, IconTrash,
} from '@tabler/icons-react';
import * as api from '../../api/labels';
import type { Label, LabelInput } from '../../api/labels';
import { LABEL_TYPES } from '../../constants/labelTypes';
import { useAuth } from '../../auth/AuthContext';
import { errorMessage } from '../../utils/errors';

const emptyForm: LabelInput = { name: '', type: '', color: '#228be6' };

function slugify(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '_')
    .replace(/^_+|_+$/g, '');
}

export function LabelsPage() {
  const { can } = useAuth();

  const [labels, setLabels] = useState<Label[]>([]);
  const [page, setPage] = useState(1);
  const [lastPage, setLastPage] = useState(1);
  const [searchInput, setSearchInput] = useState('');
  const [search] = useDebouncedValue(searchInput, 400);
  const [typeFilter, setTypeFilter] = useState<string | null>(null);
  const [showArchived, setShowArchived] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [modalOpen, setModalOpen] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [form, setForm] = useState<LabelInput>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const loadLabels = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await api.listLabels({
        page, search, type: typeFilter ?? undefined, archived: showArchived,
      });
      setLabels(result.data);
      setLastPage(result.meta.last_page);
    } catch (err) {
      setError(errorMessage(err, 'Failed to load labels.'));
    } finally {
      setLoading(false);
    }
  }, [page, search, typeFilter, showArchived]);

  useEffect(() => {
    loadLabels();
  }, [loadLabels]);

  useEffect(() => {
    setPage(1);
  }, [search, typeFilter, showArchived]);

  function openCreate() {
    setEditingId(null);
    setForm(emptyForm);
    setFormError(null);
    setModalOpen(true);
  }

  function openEdit(label: Label) {
    setEditingId(label.id);
    setForm({ name: label.name, slug: label.slug, type: label.type, color: label.color ?? '#228be6', icon: label.icon });
    setFormError(null);
    setModalOpen(true);
  }

  async function handleSave(e: FormEvent) {
    e.preventDefault();
    setSaving(true);
    setFormError(null);
    try {
      if (editingId) {
        await api.updateLabel(editingId, form);
      } else {
        await api.createLabel(form);
      }
      setModalOpen(false);
      await loadLabels();
    } catch (err) {
      setFormError(errorMessage(err, 'Failed to save.'));
    } finally {
      setSaving(false);
    }
  }

  async function handleArchive(id: number) {
    try {
      await api.archiveLabel(id);
      await loadLabels();
    } catch (err) {
      setError(errorMessage(err, 'Failed to archive.'));
    }
  }

  async function handleRestore(id: number) {
    try {
      await api.restoreLabel(id);
      await loadLabels();
    } catch (err) {
      setError(errorMessage(err, 'Failed to restore.'));
    }
  }

  async function handleForceDelete(id: number) {
    if (!window.confirm('Permanently delete this label? This cannot be undone.')) return;
    try {
      await api.forceDeleteLabel(id);
      await loadLabels();
    } catch (err) {
      setError(errorMessage(err, 'Failed to delete.'));
    }
  }

  return (
    <Stack p="xl">
      <Group justify="space-between">
        <Title order={2}>Labels</Title>
        {can('labels.create') && (
          <Button leftSection={<IconPlus size={16} />} onClick={openCreate}>
            Create
          </Button>
        )}
      </Group>

      <Group>
        <TextInput
          placeholder="Search by name"
          value={searchInput}
          onChange={(e) => setSearchInput(e.currentTarget.value)}
          w={240}
        />
        <Select
          placeholder="Filter by type"
          data={LABEL_TYPES}
          value={typeFilter}
          onChange={setTypeFilter}
          clearable
          w={220}
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

      <Table.ScrollContainer minWidth={700}>
        <Table>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Label</Table.Th>
              <Table.Th>Type</Table.Th>
              <Table.Th>Slug</Table.Th>
              <Table.Th>Actions</Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {labels.map((label) => (
              <Table.Tr key={label.id}>
                <Table.Td>
                  <Badge style={label.color ? { backgroundColor: label.color } : undefined}>
                    {label.name}
                  </Badge>
                </Table.Td>
                <Table.Td>
                  <Text size="sm">{LABEL_TYPES.find((t) => t.value === label.type)?.label ?? label.type}</Text>
                </Table.Td>
                <Table.Td>
                  <Text size="sm" c="dimmed">{label.slug}</Text>
                </Table.Td>
                <Table.Td>
                  <Group gap="xs">
                    {can('labels.edit') && !label.archived_at && (
                      <Tooltip label="Edit">
                        <ActionIcon variant="subtle" onClick={() => openEdit(label)}>
                          <IconEdit size={16} />
                        </ActionIcon>
                      </Tooltip>
                    )}
                    {can('labels.archive') && !label.archived_at && (
                      <Tooltip label="Archive">
                        <ActionIcon variant="subtle" color="orange" onClick={() => handleArchive(label.id)}>
                          <IconArchive size={16} />
                        </ActionIcon>
                      </Tooltip>
                    )}
                    {can('labels.restore') && label.archived_at && (
                      <Tooltip label="Restore">
                        <ActionIcon variant="subtle" color="green" onClick={() => handleRestore(label.id)}>
                          <IconRestore size={16} />
                        </ActionIcon>
                      </Tooltip>
                    )}
                    {can('labels.delete') && label.archived_at && (
                      <Tooltip label="Delete permanently">
                        <ActionIcon variant="subtle" color="red" onClick={() => handleForceDelete(label.id)}>
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

      {!loading && labels.length === 0 && <Text c="dimmed">No labels found.</Text>}

      <Group justify="center">
        <Pagination value={page} onChange={setPage} total={lastPage} />
      </Group>

      <Modal opened={modalOpen} onClose={() => setModalOpen(false)} title={editingId ? 'Edit label' : 'Create label'}>
        <form onSubmit={handleSave}>
          <Stack>
            {formError && (
              <Alert color="red" icon={<IconAlertTriangle size={18} />}>
                {formError}
              </Alert>
            )}
            <Select
              label="Type"
              required
              data={LABEL_TYPES}
              value={form.type || null}
              onChange={(v) => setForm((f) => ({ ...f, type: v ?? '' }))}
            />
            <TextInput
              label="Name"
              required
              value={form.name}
              onChange={(e) => setForm((f) => ({ ...f, name: e.currentTarget.value }))}
            />
            <TextInput
              label="Slug"
              description="Leave blank to auto-generate from the name"
              value={form.slug ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, slug: e.currentTarget.value }))}
              placeholder={form.name ? slugify(form.name) : undefined}
            />
            <ColorInput
              label="Color"
              required
              value={form.color}
              onChange={(color) => setForm((f) => ({ ...f, color }))}
            />
            <TextInput
              label="Icon"
              description="A Tabler icon name, e.g. IconFlag (free text — not validated against the icon library yet)"
              value={form.icon ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, icon: e.currentTarget.value }))}
            />
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
