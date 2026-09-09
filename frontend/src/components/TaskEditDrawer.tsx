import { useEffect, useRef, useState } from 'react';
import {
    Drawer, TextInput, Textarea, Select, NumberInput, Button, Group, Stack,
    Text, Avatar, ActionIcon, Divider, Progress, Checkbox, Title, NavLink, ScrollArea,
} from '@mantine/core';
import {
    IconTrash, IconPlus, IconSquareCheck, IconFolder, IconFlag,
    IconInfoCircle, IconLink, IconCalendarDollar, IconUsers, IconFileReport, IconMessageCircle, IconChartLine,
} from '@tabler/icons-react';
import type { Task, TaskInput, TaskType } from '../api/tasks';
import type { TaskDependency, TaskRelationType } from '../api/taskDependencies';
import type { InternalUser } from '../api/lookups';
import type { Label } from '../api/labels';
import { resolveIcon } from '../utils/tablerIcons';

// typeof one Tabler icon, not a generic ComponentType — @tabler/icons-react
// exports each icon as ForwardRefExoticComponent<IconProps & RefAttributes<...>>,
// which TS won't structurally match against a hand-written ComponentType
// (verified via tsc: assigning IconFolder/IconFlag to a ComponentType-typed
// field fails). Same pattern already used for this in DashboardPage.tsx.
const TASK_TYPE_OPTIONS: { id: TaskType; label: string; icon: typeof IconSquareCheck }[] = [
    { id: 'task', label: 'Task', icon: IconSquareCheck },
    { id: 'summary', label: 'Summary', icon: IconFolder },
    { id: 'milestone', label: 'Milestone', icon: IconFlag },
];

const RELATION_OPTIONS: { value: TaskRelationType; label: string }[] = [
    { value: 'FS', label: 'Finish → Start' },
    { value: 'SS', label: 'Start → Start' },
    { value: 'FF', label: 'Finish → Finish' },
    { value: 'SF', label: 'Start → Finish' },
];

const SECTIONS: { id: string; label: string; icon: typeof IconInfoCircle }[] = [
    { id: 'general', label: 'General', icon: IconInfoCircle },
    { id: 'links', label: 'Links', icon: IconLink },
    { id: 'timeline', label: 'Timeline', icon: IconCalendarDollar },
    { id: 'resources', label: 'Resources', icon: IconUsers },
    { id: 'work', label: 'Work Reports', icon: IconFileReport },
    { id: 'messages', label: 'Messages', icon: IconMessageCircle },
    { id: 'analitics', label: 'Analitics', icon: IconChartLine },
];

function getInitials(name: string) {
    return name.split(' ').filter(Boolean).slice(0, 2).map((p) => p[0].toUpperCase()).join('');
}

const formatBudget = (n: number) => `Rp ${n.toLocaleString('id-ID')}`;

interface AddDependencyFormProps {
    tasks: Task[];
    currentTaskId: number;
    onAdd: (otherTaskId: number, relation: TaskRelationType, lag: number) => Promise<void>;
}

// Both "add predecessor" and "add successor" use this — same three inputs,
// only the direction of the dependency they create differs (decided by the
// caller via onAdd, see TaskEditDrawer below).
function AddDependencyForm({ tasks, currentTaskId, onAdd }: AddDependencyFormProps) {
    const [otherId, setOtherId] = useState<string | null>(null);
    const [relation, setRelation] = useState<TaskRelationType>('FS');
    const [lag, setLag] = useState<number | ''>(0);
    const [submitting, setSubmitting] = useState(false);

    const options = tasks
        .filter((t) => t.id !== currentTaskId)
        .map((t) => ({ value: String(t.id), label: t.name }));

    async function handleAdd() {
        if (!otherId) return;
        setSubmitting(true);
        try {
            await onAdd(Number(otherId), relation, lag === '' ? 0 : lag);
            setOtherId(null);
            setLag(0);
        } finally {
            setSubmitting(false);
        }
    }

    return (
        <Group align="flex-end" gap="xs" wrap="wrap" mt="xs">
            <Select
                placeholder="Pick a predecessor task"
                data={options}
                value={otherId}
                onChange={setOtherId}
                searchable
                style={{ flex: 2, minWidth: 160 }}
            />
            <Select
                data={RELATION_OPTIONS}
                value={relation}
                onChange={(v) => setRelation((v as TaskRelationType) ?? 'FS')}
                allowDeselect={false}
                style={{ flex: 1, minWidth: 150 }}
            />
            <NumberInput
                placeholder="Lag (days)"
                value={lag}
                onChange={(v) => setLag(v === '' ? '' : Number(v))}
                style={{ width: 110 }}
            />
            <Button size="sm" leftSection={<IconPlus size={14} />} onClick={handleAdd} loading={submitting} disabled={!otherId}>
                Add
            </Button>
        </Group>
    );
}

export interface TaskEditDrawerProps {
    opened: boolean;
    task: Task | null;
    tasks: Task[];
    deps: TaskDependency[];
    internalUsers: InternalUser[];
    statusLabels: Label[];
    onClose: () => void;
    onSave: (taskId: number, input: TaskInput) => Promise<void>;
    onDelete: (taskId: number) => Promise<void>;
    onToggleCompleted: (taskId: number, completed: boolean) => Promise<void>;
    onReschedule: (taskId: number, startDate: string | null, endDate: string | null) => Promise<void>;
    onAddDependency: (sourceTaskId: number, targetTaskId: number, relation: TaskRelationType, lag: number) => Promise<void>;
    onDeleteDependency: (dependencyId: number) => Promise<void>;
}

export function TaskEditDrawer({
    opened, task, tasks, deps, internalUsers, statusLabels,
    onClose, onSave, onDelete, onToggleCompleted, onReschedule, onAddDependency, onDeleteDependency,
}: TaskEditDrawerProps) {
    const [name, setName] = useState('');
    const [description, setDescription] = useState('');
    const [type, setType] = useState<TaskType>('task');
    const [statusLabelId, setStatusLabelId] = useState<number | null>(null);
    const [assignedToUserId, setAssignedToUserId] = useState<number | null>(null);
    const [budgetPlan, setBudgetPlan] = useState<number | ''>('');
    // Separate from the rest of the form's state — start/end don't go
    // through the same onSave/TaskInput path (see handleReschedule below),
    // so tracking them together would make it look like Save covers dates
    // too, when it deliberately never does.
    const [startDate, setStartDate] = useState('');
    const [endDate, setEndDate] = useState('');
    const [saving, setSaving] = useState(false);
    const [deleting, setDeleting] = useState(false);
    const [togglingCompleted, setTogglingCompleted] = useState(false);
    const [rescheduling, setRescheduling] = useState(false);
    const [activeSection, setActiveSection] = useState(SECTIONS[0].id);
    // scrollRef = the one scrolling column (content). sectionRefs = each
    // section's own DOM node inside it, keyed by id — both read/written by
    // the scrollspy handlers below.
    const scrollRef = useRef<HTMLDivElement>(null);
    const sectionRefs = useRef<Record<string, HTMLDivElement | null>>({});



    useEffect(() => {
        if (!task) return;
        setName(task.name);
        setDescription(task.description ?? '');
        setType(task.type);
        setStatusLabelId(task.status_label?.id ?? null);
        setAssignedToUserId(task.assigned_to_user?.id ?? null);
        setBudgetPlan(task.budget_plan ?? '');
        setStartDate(task.start_date ?? '');
        setEndDate(task.end_date ?? '');
    }, [task]);

    // Every time a (possibly different) task's drawer opens, jump back to
    // the top section instead of staying wherever the previous task's
    // drawer was scrolled to.
    useEffect(() => {
        if (!opened) return;
        setActiveSection(SECTIONS[0].id);
        scrollRef.current?.scrollTo({ top: 0 });
    }, [opened, task?.id]);

    if (!task) return null;

    async function handleSubmit() {
        if (!task) return;
        setSaving(true);
        try {
            // start_date/end_date are deliberately carried over unchanged — dates
            // only move through the reschedule cascade (dragging on the chart),
            // never through this plain-field save, so a save here can never
            // silently desync a task from its dependents.
            await onSave(task.id, {
                type,
                parent_task_id: task.parent_task_id ?? null,
                name,
                description: description || null,
                assigned_to_user_id: assignedToUserId,
                start_date: task.start_date ?? null,
                end_date: task.end_date ?? null,
                budget_plan: budgetPlan === '' ? null : budgetPlan,
                budget_actual: task.budget_actual ?? null,
                weight: task.weight,
                status_label_id: statusLabelId,
            });
        } finally {
            setSaving(false);
        }
    }

    async function handleCompletedChange(checked: boolean) {
        if (!task) return;
        setTogglingCompleted(true);
        try {
            await onToggleCompleted(task.id, checked);
        } finally {
            setTogglingCompleted(false);
        }
    }

    // Deliberately its own action, not part of handleSubmit — dates go
    // through the reschedule cascade (same endpoint dragging on the chart
    // uses), which can shift dependent tasks too. Bundling it into the
    // plain field save would silently skip that cascade.
    async function handleReschedule() {
        if (!task) return;
        setRescheduling(true);
        try {
            await onReschedule(task.id, startDate || null, endDate || null);
        } finally {
            setRescheduling(false);
        }
    }

    // Fires on every scroll of the content column. A section counts as
    // "current" once its top has scrolled up to (or past) a small offset
    // below the column's own top edge — walking sections top-to-bottom and
    // keeping the last one that qualifies naturally lands on whichever
    // section the user is currently reading.
    function handleScroll() {
        const container = scrollRef.current;
        if (!container) return;
        const containerTop = container.getBoundingClientRect().top;
        let current = SECTIONS[0].id;
        for (const { id } of SECTIONS) {
            const el = sectionRefs.current[id];
            if (el && el.getBoundingClientRect().top - containerTop <= 24) {
                current = id;
            }
        }
        setActiveSection(current);
    }

    function scrollToSection(id: string) {
        sectionRefs.current[id]?.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }

    async function handleDelete() {
        if (!task || !window.confirm(`Archive "${task.name}"?`)) return;
        setDeleting(true);
        try {
            await onDelete(task.id);
        } finally {
            setDeleting(false);
        }
    }

    const taskName = (id: number) => tasks.find((t) => t.id === id)?.name ?? `#${id}`;
    const predecessors = deps.filter((d) => d.task_id === task.id);
    const createdLabel = new Date(task.created_at).toLocaleDateString('id-ID', {
        day: 'numeric', month: 'short', year: 'numeric',
    });

    const drawerTitle = (
        <Group gap="sm" wrap="nowrap" align="flex-start">
            <Checkbox
                checked={task.is_completed}
                onChange={(e) => handleCompletedChange(e.currentTarget.checked)}
                disabled={togglingCompleted}
                mt={4}
                aria-label="Mark task done"
                radius='xl'
            />
            <div>
                <Text fw={600} size="lg">{task.name}</Text>
                <Text size="xs" c="dimmed">
                    Created by {task.created_by?.name ?? 'Unknown'} on {createdLabel}
                </Text>
            </div>
        </Group>
    );

    return (
        <Drawer
            opened={opened}
            onClose={onClose}
            title={drawerTitle}
            position="right"
            size={780}
            padding="lg"
            overlayProps={{ blur: 0, opacity: 0.3 }}
            // Disable the Drawer's own outer scroll and turn its body into a
            // flex column instead — the goal is exactly ONE scrollbar (ours,
            // on the content column below), not two nested ones, and a
            // content column whose height is the real remaining space
            // (computed by flexbox) rather than a guessed vh percentage.
            styles={{
                content: { display: 'flex', flexDirection: 'column', overflow: 'hidden' },
                body: { flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column' },
            }}
        >
            {/* alignItems left at its flex default (stretch) — ScrollArea needs
                to stretch to the row's full height to have a bounded box to
                scroll within; 'flex-start' would size it to its own content
                instead, so it would never overflow and never need to scroll. */}
            <div style={{ display: 'flex', gap: 24, flex: 1, minHeight: 0 }}>
                {/* Left: one continuous scrolling column holding every section.
                    ScrollArea (not a plain overflow div) so the scrollbar is
                    a thin overlay that only appears on hover instead of a
                    permanent gutter sitting right next to the nav. */}
                <ScrollArea
                    viewportRef={scrollRef}
                    onScrollPositionChange={handleScroll}
                    type="hover"
                    style={{ flex: 1, minWidth: 0 }}
                >
                    {/* dir="rtl" (tried, to move the scrollbar left) reverted —
                        ScrollArea's internal content wrapper is display:table,
                        which reflows under rtl and shifted this whole column
                        instead of just the scrollbar. Scrollbar stays on the
                        right (ScrollArea's default) rather than fight that. */}
                    <Stack gap="lg" maw={520} mx="auto" pb="md">
                        <div ref={(el) => { sectionRefs.current.general = el; }}>
                            <Title order={5} mb="sm">General</Title>
                            <Stack gap="sm">
                                <TextInput label="Name" value={name} onChange={(e) => setName(e.currentTarget.value)} />
                                <Textarea
                                    label="Description"
                                    value={description}
                                    onChange={(e) => setDescription(e.currentTarget.value)}
                                    autosize
                                    minRows={2}
                                />
                                <Group grow align="flex-end">

                                    <Select
                                        label="Type"
                                        data={TASK_TYPE_OPTIONS.map((o) => ({ value: o.id, label: o.label }))}
                                        value={type}
                                        onChange={(v) => setType((v as TaskType) ?? 'task')}
                                        allowDeselect={false}
                                        leftSection={(() => {
                                            const Icon = TASK_TYPE_OPTIONS.find((o) => o.id === type)?.icon;
                                            return Icon ? <Icon size={16} /> : null;
                                        })()}
                                        renderOption={({ option }) => {
                                            const opt = TASK_TYPE_OPTIONS.find((o) => o.id === option.value);
                                            const Icon = opt?.icon;
                                            return (
                                                <Group gap={8} wrap="nowrap">
                                                    {Icon && <Icon size={16} />}
                                                    <Text size="sm">{opt?.label}</Text>
                                                </Group>
                                            );
                                        }}
                                    />
                                    <Select
                                        label="Status"
                                        placeholder="No status"
                                        clearable
                                        data={statusLabels.map((l) => ({ value: String(l.id), label: l.name }))}
                                        value={statusLabelId != null ? String(statusLabelId) : null}
                                        onChange={(v) => setStatusLabelId(v ? Number(v) : null)}
                                        leftSection={(() => {
                                            const label = statusLabels.find((l) => l.id === statusLabelId);
                                            const Icon = resolveIcon(label?.icon);
                                            return Icon ? <Icon size={16} /> : null;
                                        })()}
                                        renderOption={({ option }) => {
                                            const label = statusLabels.find((l) => String(l.id) === option.value);
                                            const Icon = resolveIcon(label?.icon);
                                            return (
                                                <Group gap={8} wrap="nowrap">
                                                    {Icon && <Icon size={16} />}
                                                    <Text size="sm">{label?.name}</Text>
                                                </Group>
                                            );
                                        }}
                                    />
                                </Group>
                                <Select
                                    label="Assignee"
                                    placeholder="Unassigned"
                                    clearable
                                    searchable
                                    data={internalUsers.map((u) => ({ value: String(u.id), label: u.name }))}
                                    value={assignedToUserId != null ? String(assignedToUserId) : null}
                                    onChange={(v) => setAssignedToUserId(v ? Number(v) : null)}
                                    leftSection={(() => {
                                        const user = internalUsers.find((u) => u.id === assignedToUserId);
                                        if (!user) return null;
                                        return (
                                            <Avatar size={20} radius="xl" color="blue" variant="light" src={user.avatar}>
                                                {getInitials(user.name)}
                                            </Avatar>
                                        );
                                    })()}
                                    renderOption={({ option }) => {
                                        const user = internalUsers.find((u) => String(u.id) === option.value);
                                        if (!user) return null;
                                        return (
                                            <Group gap={8} wrap="nowrap">
                                                <Avatar size={20} radius="xl" color="blue" variant="light" src={user.avatar}>
                                                    {getInitials(user.name)}
                                                </Avatar>
                                                <Text size="sm">{user.name}</Text>
                                            </Group>
                                        );
                                    }}
                                />
                            </Stack>
                        </div>

                        <div ref={(el) => { sectionRefs.current.links = el; }}>
                            <Title order={5} mb="sm">Links</Title>
                            <Stack gap="lg">
                                <div>
                                    <Text fw={500} size="sm">Predecessors</Text>
                                    <Text size="xs" c="dimmed" mb={8}>Tasks this task depends on.</Text>
                                    {predecessors.length === 0 ? (
                                        <Text size="sm" c="dimmed">None yet</Text>
                                    ) : (
                                        <Stack gap={4}>
                                            {predecessors.map((d) => (
                                                <Group key={d.id} justify="space-between" wrap="nowrap">
                                                    <Text size="sm">
                                                        {taskName(d.depends_on_task_id)} · {d.relation_type}
                                                        {d.lag_days ? ` · lag ${d.lag_days}d` : ''}
                                                    </Text>
                                                    <ActionIcon variant="subtle" color="red" onClick={() => onDeleteDependency(d.id)}>
                                                        <IconTrash size={14} />
                                                    </ActionIcon>
                                                </Group>
                                            ))}
                                        </Stack>
                                    )}
                                    <AddDependencyForm
                                        tasks={tasks}
                                        currentTaskId={task.id}
                                        onAdd={(otherId, relation, lag) => onAddDependency(otherId, task.id, relation, lag)}
                                    />
                                </div>

                                <Divider />
                            </Stack>
                        </div>

                        <div ref={(el) => { sectionRefs.current.timeline = el; }}>
                            <Title order={5} mb="sm">Timeline &amp; Budget</Title>
                            <Stack gap="sm">
                                <Group grow align="flex-end">
                                    <TextInput
                                        type="date"
                                        label="Start date"
                                        value={startDate}
                                        onChange={(e) => setStartDate(e.currentTarget.value)}
                                    />
                                    <TextInput
                                        type="date"
                                        label="End date"
                                        value={endDate}
                                        onChange={(e) => setEndDate(e.currentTarget.value)}
                                    />
                                </Group>
                                <Group justify="space-between" align="center">
                                    <Text size="xs" c="dimmed" style={{ flex: 1 }}>
                                        Changing dates here goes through the same reschedule cascade as dragging on the
                                        chart — dependent tasks may shift too.
                                    </Text>
                                    <Button
                                        size="xs"
                                        variant="light"
                                        onClick={handleReschedule}
                                        loading={rescheduling}
                                        disabled={startDate === (task.start_date ?? '') && endDate === (task.end_date ?? '')}
                                    >
                                        Reschedule
                                    </Button>
                                </Group>
                                <NumberInput
                                    label="Budget plan"
                                    value={budgetPlan}
                                    onChange={(v) => setBudgetPlan(v === '' ? '' : Number(v))}
                                    prefix="Rp "
                                    thousandSeparator="."
                                    decimalScale={0}
                                    decimalSeparator=','
                                    min={0}
                                />
                                <Text size="sm" c="dimmed">
                                    Actual: {task.budget_actual != null ? formatBudget(task.budget_actual) : '—'}
                                </Text>
                            </Stack>
                        </div>

                        <div ref={(el) => { sectionRefs.current.resources = el; }}>
                            <Title order={5} mb="sm">Resources</Title>
                            <Text c="dimmed" size="sm">This feature isn't built yet — coming in a future update.</Text>
                        </div>

                        <div ref={(el) => { sectionRefs.current.work = el; }}>
                            <Title order={5} mb="sm">Work Reports</Title>
                            <Text c="dimmed" size="sm">This feature isn't built yet — coming in a future update.</Text>
                        </div>

                        <div ref={(el) => { sectionRefs.current.messages = el; }}>
                            <Title order={5} mb="sm">Messages</Title>
                            <Text c="dimmed" size="sm">This feature isn't built yet — coming in a future update.</Text>
                        </div>

                        <div ref={(el) => { sectionRefs.current.analitics = el; }}>
                            <Title order={5} mb="sm">Analitics</Title>
                            <Stack gap={4}>
                                <Text size="sm" fw={600}>Progress</Text>
                                <Progress value={task.progress ?? 0} size="lg" />
                                <Text size="xs" c="dimmed">
                                    {Math.round(task.progress ?? 0)}% — derived from cumulative work reports, not editable here.
                                </Text>
                            </Stack>
                        </div>

                        <Group justify="space-between" pt="md">
                            <Button variant="subtle" color="red" leftSection={<IconTrash size={14} />} onClick={handleDelete} loading={deleting}>
                                Archive task
                            </Button>
                            <Group>
                                <Button variant="default" onClick={onClose}>Cancel</Button>
                                <Button onClick={handleSubmit} loading={saving}>Save</Button>
                            </Group>
                        </Group>
                    </Stack>
                </ScrollArea>

                {/* Right: sticky-feeling nav — it's simply outside the scrolling
                    column above, so it never moves while that column scrolls;
                    no CSS position:sticky needed. */}
                <Stack gap={2} style={{ width: 150, flexShrink: 0 }}>
                    {SECTIONS.map((s) => (
                        <NavLink
                            key={s.id}
                            label={s.label}
                            active={activeSection === s.id}
                            onClick={() => scrollToSection(s.id)}
                            variant="light"
                            style={{ borderRadius: 6 }}
                            leftSection={<s.icon size={16} />}
                        />
                    ))}
                </Stack>
            </div>
        </Drawer>
    );
}
