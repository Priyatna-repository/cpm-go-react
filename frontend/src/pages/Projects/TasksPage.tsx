import { useEffect, useState, useRef } from 'react';
import { Link, useParams } from 'react-router-dom';
import { Anchor, Breadcrumbs, Stack, Text, Title, Center } from '@mantine/core';
import { useLayoutChrome } from '../../layouts/LayoutChromeContext';
import * as projectsApi from '../../api/projects';
import type { Project } from '../../api/projects';
import { GanttPanel } from '../../components/GanttPanel';
import { errorMessage } from '../../utils/errors';

export function TasksPage() {
  const { id } = useParams<{ id: string }>();
  const projectId = Number(id);
  const { fullscreen } = useLayoutChrome();

  const [project, setProject] = useState<Project | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [ganttHeight, setGanttHeight] = useState<number | null>(null);
  const ganttContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    projectsApi
      .getProject(projectId)
      .then(setProject)
      .catch((err) => setError(errorMessage(err, 'Failed to load project')));
  }, [projectId]);

  function handleResizeStart(e: React.MouseEvent){
    e.preventDefault();
    const startY = e.clientY; //ukuran tinggi sekarnag langsung di DOM bukan di hitung ulang rumus
    const startHeight = ganttContainerRef.current?.getBoundingClientRect().height ?? 400;

  function onMouseMove(ev: MouseEvent){
    const delta = ev.clientY - startY;
    setGanttHeight(Math.max(300, startHeight + delta));
  }

  function onMouseUp(){
    window.removeEventListener('mousemove', onMouseMove);
    window.removeEventListener('mouseup', onMouseUp);
  }

  window.addEventListener('mousemove', onMouseMove);
  window.addEventListener('mouseup', onMouseUp);
}

  return (
    <Stack
      gap="lg"
      style={{
        height:
          ganttHeight == null
            ? 'calc(100dvh - var(--app-shell-header-offset, 0rem) - var(--app-shell-footer-offset, 0rem) - (var(--app-shell-padding, 0rem) * 2))'
            : undefined,
      }}
    >
      {error && <Text c="red">{error}</Text>}
      { fullscreen ? (
        <Center py={'sm'}>
          <Title order={4}>{project?.name ?? '...'}</Title>
        </Center>
      ):(
        <Breadcrumbs mb="xs">
        <Anchor component={Link} to="/projects" size="sm" c="dimmed">Projects</Anchor>
        <Text size="sm">{project?.name ?? '…'}</Text>
      </Breadcrumbs>
      )}
      <div
        ref={ganttContainerRef}
        style={ganttHeight != null ? { height: ganttHeight } : { flex: 1, minHeight: 0 }}
      >
        <GanttPanel projectId={projectId} />
      </div>
    <div
        onMouseDown={handleResizeStart}
        style={{
          flexShrink: 0,
          height: 10,
          cursor: 'ns-resize',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <div
          style={{
            width: 48,
            height: 4,
            borderRadius: 2,
            backgroundColor: 'var(--mantine-color-gray-4)',
          }}
        />
      </div>
    </Stack>
  )};
