import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Gantt, Toolbar, Willow } from '@svar-ui/react-gantt';
import '@svar-ui/react-gantt/all.css';
import { Alert, Text, Avatar, Badge, Group, Tooltip, Stack } from '@mantine/core';
import * as tasksApi from '../api/tasks';
import type { Task, TaskType, TaskInput } from '../api/tasks';
import * as depsApi from '../api/taskDependencies';
import type { TaskDependency, TaskRelationType } from '../api/taskDependencies';
import * as lookupsApi from '../api/lookups';
import type { InternalUser } from '../api/lookups';
import * as labelsApi from '../api/labels';
import type { Label } from '../api/labels';
import { errorMessage } from '../utils/errors';
import { resolveIcon } from '../utils/tablerIcons';
import { TaskEditDrawer } from './TaskEditDrawer';

const RELATION_TO_LINK_TYPE: Record<TaskRelationType, string> = {
  FS: 'e2s',
  SS: 's2s',
  FF: 'e2e',
  SF: 's2e',
};
const LINK_TYPE_TO_RELATION: Record<string, TaskRelationType> = {
  e2s: 'FS',
  s2s: 'SS',
  e2e: 'FF',
  s2e: 'SF',
};

interface SvarTask {
  id: number;
  text: string;
  start?: Date;
  end?: Date;
  duration?: number;
  progress?: number;
  type: TaskType;
  parent: number;
  // Extra fields beyond SVAR's own ITask shape — it allows arbitrary extra
  // keys ([key: string]: any). These carry the real backend ids (not
  // display strings) so the SAME field the grid's `getter` reads for
  // display is also what the Editor's select fields read/write — one
  // source of truth instead of a display copy the form can't touch.
  assigned_to_user_id?: number;
  status_label_id?: number;
  budget_plan?: number;
}

interface SvarLink {
  id: number;
  source: number;
  target: number;
  type: string;
  lag?: number;
}

function toSvarData(tasks: Task[], deps: TaskDependency[]) {
  // SVAR requires a summary row to either have subtasks or its own explicit
  // start/end. A freshly created summary task with neither yet (before any
  // children or dates are added) would crash the chart, so it's left out
  // until it has one or the other — it's still fully editable via the
  // Editor form, just not drawn on the timeline yet.
  const hasChildren = (id: number) => tasks.some((t) => t.parent_task_id === id);
  const renderable = tasks.filter((t) => t.type !== 'summary' || hasChildren(t.id) || (t.start_date && t.end_date));

  const svarTasks: SvarTask[] = renderable.map((t) => ({
    id: t.id,
    text: t.name,
    start: t.start_date ? new Date(t.start_date) : undefined,
    end: t.end_date ? new Date(t.end_date) : undefined,
    progress: t.progress,
    type: t.type,
    parent: t.parent_task_id ?? 0,
    assigned_to_user_id: t.assigned_to_user?.id,
    status_label_id: t.status_label?.id,
    budget_plan: t.budget_plan,
  }));
  const svarLinks: SvarLink[] = deps.map((d) => ({
    id: d.id,
    source: d.depends_on_task_id,
    target: d.task_id,
    type: RELATION_TO_LINK_TYPE[d.relation_type],
    lag: d.lag_days,
  }));
  return { svarTasks, svarLinks };
}

// Mirrors what the reference app's TaskRow.jsx shows per task: assignee,
// Start, End, dependency relation, Progress, Status. (Reference also shows
// Priority — we don't have a priority field on Task yet, left out rather
// than faked; add task.priority_label_id later, same shape as
// status_label_id, if that's wanted.)
//
// Both this and the Editor's items (built in the component below) resolve
// display names from the same internalUsers/statusLabels lookup lists, so
// a column and its corresponding form field always agree.
// Hardcoded to Rupiah — Project doesn't currently expose its real
// configured currency to the frontend (that lives on
// ClientCompany.currency_id, one level removed), so this is a stand-in
// assumption matching the app's context, not something derived per
// project. Revisit if a project ever needs a different currency shown.
const formatBudget = (n: number) => `Rp ${n.toLocaleString('id-ID')}`;

function getInitials(name: string) {
  return name
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0].toUpperCase())
    .join('');
}

function computeWbsMap(tasks: SvarTask[]): Map<number, string> {
  const childrenByParent = new Map<number, SvarTask[]>();
  for (const t of tasks) {
    const parentId = t.parent || 0;
    if (!childrenByParent.has(parentId)) childrenByParent.set(parentId, []);
    childrenByParent.get(parentId)!.push(t);
  }
  const wbsById = new Map<number, string>();
  function assign(parentId: number, prefix: string) {
    const siblings = childrenByParent.get(parentId) ?? [];
    siblings.forEach((t, i) => {
      const code = prefix ? `${prefix}.${i + 1}` : `${i + 1}`;
      wbsById.set(t.id, code);
      assign(t.id, code);
    });
  }
  assign(0, '');
  return wbsById;
}

function buildColumns(internalUsers: InternalUser[], statusLabels: Label[], svarTasks: SvarTask[], svarLinks: SvarLink[]) {
  const taskNameById = new Map(svarTasks.map((t) => [t.id, t.text]));
  const PredecessorsCell = ({ row }: { row: any }) => {
    const preds = svarLinks
      .filter((l) => l.target === row.id)
      .map((l) => ({
        name: taskNameById.get(l.source) ?? `#${l.source}`,
        relation: LINK_TYPE_TO_RELATION[l.type] ?? l.type,
      }));
    if (preds.length === 0) return <Text size="xs" c="dimmed">—</Text>;
    return (
      <Tooltip
        label={
          <Stack gap={2}>
            {preds.map((p, i) => (
              <Text key={i} size="xs">{p.name} ({p.relation})</Text>
            ))}
          </Stack>
        }
        position="top"
        withArrow
      >
        <Text size="xs">{preds.map((p) => p.relation).join(', ')}</Text>
      </Tooltip>
    );
  };
  const AssigneeCell = ({ row }: { row: any }) => {
    const user = internalUsers.find((u) => u.id === row.assigned_to_user_id);
    if (!user) return <Text size="xs" c="dimmed">—</Text>;
    return (
      <Group gap={6} wrap="nowrap" py={4}>
        <Tooltip label={user.name} position="top" withArrow>
          <Avatar
            size={22}
            radius="xl"
            color="blue"
            variant="light"
            src={user.avatar ? `${import.meta.env.VITE_API_BASE_URL}${user.avatar}` : undefined}
          >
            {getInitials(user.name)}
          </Avatar>
        </Tooltip>
      </Group>
    );
  };
  const StatusCell = ({ row }: { row: any }) => {
    const label = statusLabels.find((l) => l.id === row.status_label_id);
    if (!label) return <Text size="xs" c="dimmed">—</Text>;
    const StatusIcon = resolveIcon(label.icon);
    return (
      <Badge
        size="sm"
        leftSection={StatusIcon ? <StatusIcon size={12} /> : undefined}
        style={label.color ? { backgroundColor: label.color, color: '#fff' } : undefined}
        styles={{ label: { textTransform: 'none' } }}
      >
        {label.name}
      </Badge>
    );
  };
  const wbsById = computeWbsMap(svarTasks);
  return [
    { id: 'wbs', header: '#', align: 'center' as const, width: 70, getter: (t: any) => wbsById.get(t.id) ?? '' },
    { id: 'text', header: 'Task name' },
    { id: 'assigned_to_user_id', header: 'Assignee', align: 'center' as const, cell: AssigneeCell },
    { id: 'start', header: 'Start', align: 'center' as const },
    { id: 'end', header: 'End', align: 'center' as const },
    { id: 'duration', header: 'Duration', align: 'center' as const },
    { id: 'progress', header: 'Progress', align: 'center' as const, getter: (t: any) => (t.progress != null ? `${Math.round(t.progress)}%` : '—'), },
    { id: 'status_label_id', header: 'Status', align: 'center' as const, cell: StatusCell },
    { id: 'budget_plan', header: 'Budget', align: 'center' as const, getter: (t: any) => (t.budget_plan != null ? formatBudget(t.budget_plan) : '—'), },
    { id: 'predecessors', header: 'Predecessors', align: 'center' as const, cell: PredecessorsCell },
  ];
}

export function GanttPanel({ projectId }: { projectId: number }) {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [svarTasks, setSvarTasks] = useState<SvarTask[]>([]);
  const [svarLinks, setSvarLinks] = useState<SvarLink[]>([]);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [deps, setDeps] = useState<TaskDependency[]>([]);
  // Which task's edit drawer is open — null means closed.
  const [editingTaskId, setEditingTaskId] = useState<number | null>(null);

  // Populate the Assignee/Status dropdowns in both the grid columns and the
  // Editor form. Fetched separately from load() and failures are swallowed
  // — listInternalUsers requires project.manage_access, which Team
  // Members/Clients viewing this same chart won't have, and that's fine:
  // the columns/fields just show ids with no name resolved rather than
  // breaking the whole page over a lookup only Managers can use anyway.
  const [internalUsers, setInternalUsers] = useState<InternalUser[]>([]);
  const [statusLabels, setStatusLabels] = useState<Label[]>([]);
  // Keyed by our own numeric task id — used to fill in whichever of
  // start/end/text/type a partial event payload didn't include.
  const tasksByIdRef = useRef<Map<number, Task>>(new Map());
  // apiRef is for imperative calls inside handlers (apiRef.current.exec(...));
  // ganttApi (state) is for the `api` prop Toolbar/Editor expect — they want
  // the resolved IApi instance itself, not a ref wrapper, so a plain useRef
  // doesn't satisfy their prop type. The Gantt ref callback below sets both.
  const apiRef = useRef<any>(null);
  const [ganttApi, setGanttApi] = useState<any>(null);

  // Must be stable (useCallback, not an inline arrow in the JSX below) — an
  // inline function is a new identity every render, so React would detach
  // and reattach the ref on every render, and reattaching calls
  // setGanttApi(), triggering another render: an infinite loop.
  const handleGanttRef = useCallback((el: any) => {
    apiRef.current = el;
    setGanttApi(el);
  }, []);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [fetchedTasks, fetchedDeps] = await Promise.all([
        tasksApi.listTasks(projectId),
        depsApi.listTaskDependencies(projectId),
      ]);
      tasksByIdRef.current = new Map(fetchedTasks.map((t) => [t.id, t]));
      setTasks(fetchedTasks);
      setDeps(fetchedDeps);
      const { svarTasks: st, svarLinks: sl } = toSvarData(fetchedTasks, fetchedDeps);
      setSvarTasks(st);
      setSvarLinks(sl);
    } catch (err) {
      setError(errorMessage(err, 'Failed to load the Gantt chart'));
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    lookupsApi.listInternalUsers().then(setInternalUsers).catch(() => { });
    labelsApi.listLabels({ type: 'pt_status' }).then((r) => setStatusLabels(r.data)).catch(() => { });
  }, []);

  // Double-click on a row/bar, or the Toolbar's Edit button, already fires
  // 'show-editor' on the shared api — verified straight from SVAR's
  // DataStore source: it's baked in there, not owned by the <Editor>
  // component we just removed. Subscribing here is all that's needed to
  // keep exactly the same open-drawer triggers as before.
  useEffect(() => {
    if (!ganttApi) return;
    const handler = (ev: { id: number | string | null }) => {
      setEditingTaskId(ev.id != null ? Number(ev.id) : null);
    };
    ganttApi.on('show-editor', handler, { tag: 'task-edit-drawer' });
    return () => ganttApi.detach('task-edit-drawer');
  }, [ganttApi]);

  const columns = useMemo(
    () => buildColumns(internalUsers, statusLabels, svarTasks, svarLinks),
    [internalUsers, statusLabels, svarTasks, svarLinks],
  );

  // The Drawer isn't part of SVAR's own drag/Toolbar event flow, so it
  // persists straight through the REST API and reloads — same pattern
  // handleAddTask/handleDeleteTask below already use.
  const handleDrawerSave = async (taskId: number, input: TaskInput) => {
    try {
      await tasksApi.updateTask(projectId, taskId, input);
      setEditingTaskId(null);
      await load();
    } catch (err) {
      setError(errorMessage(err, 'Failed to save changes'));
    }
  };

    const handleToggleCompleted = async (taskId: number, completed: boolean) => {
    try {
      await tasksApi.setTaskCompleted(projectId, taskId, completed);
      await load();
    } catch (err) {
      setError(errorMessage(err, 'Failed to update completion status'));
    }
  };

  // Drawer-driven date edits — same cascade endpoint as chart-drag
  // (handleUpdateTask's date branch below), but called directly since this
  // isn't a drag gesture: no optimistic per-task exec() pushback needed,
  // just persist and reload everything fresh.
  const handleDrawerReschedule = async (taskId: number, startDate: string | null, endDate: string | null) => {
    try {
      await tasksApi.rescheduleTask(projectId, taskId, startDate, endDate);
      await load();
    } catch (err) {
      setError(errorMessage(err, 'Failed to reschedule'));
    }
  };

  // Fires on every field change: dragging a bar (dates), editing the name
  // or type via the built-in Editor form. Date changes go through the
  // reschedule cascade; everything else is a plain field update.
  // inProgress ticks (mid-drag) are ignored — the reschedule endpoint does
  // real work that shouldn't run on every pixel of mouse movement.
  const handleUpdateTask = async (ev: {
    id: string | number;
    task: Partial<{
      text: string;
      type: string;
      parent: string | number;
      start: Date;
      end: Date;
      duration: number;
      progress: number;
      assigned_to_user_id: number | null;
      status_label_id: number | null;
      budget_plan: number | string | null;
    }>;
    inProgress?: boolean;
  }) => {
    if (ev.inProgress) return;
    const taskId = Number(ev.id);
    const current = tasksByIdRef.current.get(taskId);
    if (!current) return;

    const datesChanged = ev.task.start !== undefined || ev.task.end !== undefined || ev.task.duration !== undefined;
    if (datesChanged) {
      const newStart = ev.task.start ?? (current.start_date ? new Date(current.start_date) : undefined);
      let newEnd = ev.task.end;
      if (!newEnd && ev.task.duration && newStart) {
        newEnd = new Date(newStart.getTime() + ev.task.duration * 24 * 60 * 60 * 1000);
      }
      newEnd = newEnd ?? (current.end_date ? new Date(current.end_date) : undefined);
      if (!newStart || !newEnd) return;

      const toDateOnly = (d: Date) => d.toISOString().slice(0, 10);
      try {
        const changed = await tasksApi.rescheduleTask(projectId, taskId, toDateOnly(newStart), toDateOnly(newEnd));
        // Push every affected task's new dates back into the Gantt's own
        // store — it already applied the dragged task's own change
        // locally, but has no idea the cascade also moved its dependents.
        for (const t of changed) {
          tasksByIdRef.current.set(t.id, t);
          if (t.id === taskId) continue; // already reflected locally by the drag itself
          apiRef.current?.exec('update-task', {
            id: t.id,
            task: { start: t.start_date ? new Date(t.start_date) : undefined, end: t.end_date ? new Date(t.end_date) : undefined },
          });
        }
      } catch (err) {
        setError(errorMessage(err, 'Failed to reschedule — reverting'));
        await load(); // reload from the server so the chart matches reality, undoing the optimistic drag
      }
      return;
    }

    // A non-date field changed (name, type, parent, ...) — typically via
    // the built-in Editor form. Merge onto what we know and persist as a
    // regular update, not a reschedule.
    const newParent = ev.task.parent !== undefined ? (Number(ev.task.parent) || null) : (current.parent_task_id ?? null);
    try {
      const assignedToUserId =
        ev.task.assigned_to_user_id !== undefined ? ev.task.assigned_to_user_id : (current.assigned_to_user?.id ?? null);
      const statusLabelId =
        ev.task.status_label_id !== undefined ? ev.task.status_label_id : (current.status_label?.id ?? null);
      // The text field's raw value may include thousand separators a user
      // types naturally (e.g. "150.000.000" or "150,000,000") — strip
      // everything but digits before parsing, rather than rejecting input
      // that isn't a bare number.
      let budgetPlan: number | null;
      if (ev.task.budget_plan === undefined) {
        budgetPlan = current.budget_plan ?? null;
      } else if (ev.task.budget_plan === null || ev.task.budget_plan === '') {
        budgetPlan = null;
      } else if (typeof ev.task.budget_plan === 'string') {
        const digits = ev.task.budget_plan.replace(/[^0-9]/g, '');
        budgetPlan = digits ? Number(digits) : null;
      } else {
        budgetPlan = ev.task.budget_plan;
      }

      await tasksApi.updateTask(projectId, taskId, {
        type: (ev.task.type as TaskType) ?? current.type,
        parent_task_id: newParent,
        name: ev.task.text ?? current.name,
        description: current.description,
        assigned_to_user_id: assignedToUserId,
        start_date: current.start_date,
        end_date: current.end_date,
        budget_plan: budgetPlan,
        budget_actual: current.budget_actual,
        weight: current.weight,
        status_label_id: statusLabelId,
      });
      tasksByIdRef.current.set(taskId, { ...current, name: ev.task.text ?? current.name });
    } catch (err) {
      setError(errorMessage(err, 'Failed to save changes'));
      await load();
    }
  };

  // Fires when a new row is added via the Toolbar's "Add task" button —
  // defaults to a plain task; change its type to "summary" or "milestone"
  // afterward via the Editor form.
  const handleAddTask = async (ev: { task: Partial<{ text: string; type: string; parent: string | number }> }) => {
    const parentId = ev.task.parent !== undefined && Number(ev.task.parent) !== 0 ? Number(ev.task.parent) : null;
    const type = (ev.task.type as TaskType) || 'task';
    try {
      await tasksApi.createTask(projectId, { type, parent_task_id: parentId, name: ev.task.text || 'New task', weight: 0 });
    } catch (err) {
      setError(errorMessage(err, 'Failed to create task'));
    } finally {
      await load();
    }
  };

  // Fires from the Editor's delete action — maps to our soft Archive, not
  // a permanent delete.
  const handleDeleteTask = async (ev: { id: string | number }) => {
    const taskId = Number(ev.id);
    try {
      await tasksApi.archiveTask(projectId, taskId);
    } catch (err) {
      setError(errorMessage(err, 'Failed to archive task'));
    } finally {
      await load();
    }
  };

  // Fires when the user drags from one bar's edge to another to draw a new
  // dependency. The library adds the link to its own store optimistically
  // before we hear about it — so on rejection (e.g. a cycle) we must
  // explicitly remove it again, not just show an error.
  const handleAddLink = async (ev: {
    id?: string | number;
    link: { id?: string | number; source?: string | number; target?: string | number; type?: string; lag?: number };
  }) => {
    if (ev.link.target === undefined || ev.link.source === undefined || !ev.link.type) return;
    const successorId = Number(ev.link.target);
    const predecessorId = Number(ev.link.source);
    const relation = LINK_TYPE_TO_RELATION[ev.link.type];
    const linkId = ev.id ?? ev.link.id;
    if (!relation) return;

    try {
      await depsApi.createTaskDependency(projectId, successorId, predecessorId, relation, ev.link.lag ?? 0);
      await load(); // pick up the real dependency id assigned by the backend
    } catch (err) {
      setError(errorMessage(err, 'Could not create that dependency'));
      if (linkId !== undefined) {
        apiRef.current?.exec('delete-link', { id: linkId });
      }
    }
  };

  const handleDeleteLink = async (ev: { id: string | number }) => {
    try {
      await depsApi.deleteTaskDependency(projectId, Number(ev.id));
    } catch (err) {
      setError(errorMessage(err, 'Failed to delete dependency'));
      await load();
    }
  };

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      {error && (
        <Alert color="red" onClose={() => setError(null)} withCloseButton mb="xs">
          {error}
        </Alert>
      )}

      {loading ? (
        <Text c="dimmed" p="md">Loading Gantt chart…</Text>
      ) : svarTasks.length === 0 ? (
        <Text c="dimmed" p="md">No tasks yet — use the "Add task" button below to create one.</Text>
      ) : null}

      <div
        style={{
          flex: 1,
          minHeight: 0,
          display: loading ? 'none' : 'block',
          border: '1px solid var(--mantine-color-gray-4)',
          borderRadius: 'var(--mantine-radius-sm)',
          overflow: 'hidden',
        }}
      >
        <Willow>
          {/* Toolbar/Editor don't tolerate a null `api` on first mount — they
              assume a resolved instance synchronously. Since they're siblings
              of Gantt, they'd otherwise render once (and crash) before
              Gantt's ref callback ever fires. The `ganttApi &&` guard defers
              their first mount to the render pass right after Gantt has
              already reported its api instance — JSX position (not mount
              order) still controls visual placement, so Toolbar stays above
              the chart despite mounting slightly later. */}
          {ganttApi && <Toolbar api={ganttApi} />}
          <div style={{ borderTop: '1px solid var(--mantine-color-gray-4)' }}>
            <Gantt
              ref={handleGanttRef}
              tasks={svarTasks}
              links={svarLinks}
              columns={columns}
              onupdatetask={handleUpdateTask}
              onaddtask={handleAddTask}
              ondeletetask={handleDeleteTask}
              onaddlink={handleAddLink}
              ondeletelink={handleDeleteLink}
            />
          </div>
          <TaskEditDrawer
            opened={editingTaskId != null}
            task={editingTaskId != null ? tasks.find((t) => t.id === editingTaskId) ?? null : null}
            tasks={tasks}
            deps={deps}
            internalUsers={internalUsers}
            statusLabels={statusLabels}
            onClose={() => setEditingTaskId(null)}
            onSave={handleDrawerSave}
            onReschedule={handleDrawerReschedule}
            onToggleCompleted={handleToggleCompleted}
            onDelete={async (id) => { await handleDeleteTask({ id }); setEditingTaskId(null); }}
            onAddDependency={(source, target, relation, lag) =>
              handleAddLink({ link: { source, target, type: RELATION_TO_LINK_TYPE[relation], lag } })
            }
            onDeleteDependency={(id) => handleDeleteLink({ id })}
          />

        </Willow>
      </div>
    </div>
  );
}
