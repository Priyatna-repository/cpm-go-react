import { useEffect, useState } from 'react';
import { Button, Group, Paper, SimpleGrid, Skeleton, Stack, Text, Title } from '@mantine/core';
import { Link } from 'react-router-dom';
import { IconBuildingSkyscraper, IconFolders, IconPlus } from '@tabler/icons-react';
import { useAuth } from '../../auth/AuthContext';
import * as projectsApi from '../../api/projects';
import type { Project } from '../../api/projects';
import * as clientCompaniesApi from '../../api/clientCompanies';

interface StatTile {
  label: string;
  value: number | null;
  icon: typeof IconFolders;
  color: string;
  to: string;
}

export function DashboardPage() {
  const { user, can } = useAuth();
  const [projectCount, setProjectCount] = useState<number | null>(null);
  const [companyCount, setCompanyCount] = useState<number | null>(null);
  const [recentProjects, setRecentProjects] = useState<Project[]>([]);

  useEffect(() => {
    (async () => {
      if (can('project.view')) {
        try {
          const result = await projectsApi.listProjects({ page: 1 });
          setProjectCount(result.meta.total);
          setRecentProjects(result.data.slice(0, 5));
        } catch {
          setProjectCount(null);
        }
      }
      if (can('client_company.view')) {
        try {
          const result = await clientCompaniesApi.listClientCompanies({ page: 1 });
          setCompanyCount(result.meta.total);
        } catch {
          setCompanyCount(null);
        }
      }
    })();
    // Re-fetch when the logged-in user changes (login/logout), not on every render.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user]);

  const tiles: StatTile[] = [];
  if (can('project.view')) {
    tiles.push({ label: 'Projects', value: projectCount, icon: IconFolders, color: 'blue', to: '/projects' });
  }
  if (can('client_company.view')) {
    tiles.push({
      label: 'Client Companies',
      value: companyCount,
      icon: IconBuildingSkyscraper,
      color: 'grape',
      to: '/clients/companies',
    });
  }

  return (
    <Stack gap="xl">
      <div>
        <Title order={2}>Welcome, {user?.name}</Title>
        <Text c="dimmed" size="sm">
          Logged in as {user?.role}
        </Text>
      </div>

      {tiles.length > 0 && (
        <SimpleGrid cols={{ base: 1, sm: 2, md: 3 }}>
          {tiles.map((tile) => (
            <Paper
              key={tile.label}
              component={Link}
              to={tile.to}
              withBorder
              p="md"
              radius="md"
              style={{ textDecoration: 'none' }}
            >
              <Group justify="space-between">
                <div>
                  <Text size="xs" c="dimmed" tt="uppercase" fw={700}>
                    {tile.label}
                  </Text>
                  {tile.value === null ? (
                    <Skeleton height={28} width={50} mt={4} />
                  ) : (
                    <Text fw={700} size="xl">
                      {tile.value}
                    </Text>
                  )}
                </div>
                <tile.icon size={28} color={`var(--mantine-color-${tile.color}-6)`} stroke={1.5} />
              </Group>
            </Paper>
          ))}
        </SimpleGrid>
      )}

      <Group>
        {can('project.create') && (
          <Button leftSection={<IconPlus size={16} />} component={Link} to="/projects" variant="light">
            New Project
          </Button>
        )}
        {can('client_company.create') && (
          <Button
            leftSection={<IconPlus size={16} />}
            component={Link}
            to="/clients/companies"
            variant="light"
            color="grape"
          >
            New Client Company
          </Button>
        )}
      </Group>

      {can('project.view') && recentProjects.length > 0 && (
        <div>
          <Title order={4} mb="sm">
            Recent Projects
          </Title>
          <Stack gap="xs">
            {recentProjects.map((p) => (
              <Paper key={p.id} component={Link} to="/projects" withBorder p="sm" radius="md" style={{ textDecoration: 'none' }}>
                <Group justify="space-between">
                  <div>
                    <Text fw={500} size="sm">
                      {p.name}
                    </Text>
                    <Text size="xs" c="dimmed" ff="monospace">
                      {p.code}
                    </Text>
                  </div>
                  <Text size="xs" c="dimmed">
                    {p.client_company?.name ?? p.client_user?.name ?? '—'}
                  </Text>
                </Group>
              </Paper>
            ))}
          </Stack>
        </div>
      )}
    </Stack>
  );
}
