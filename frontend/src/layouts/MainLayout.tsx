import { AppShell, Avatar, Burger, Group, Menu, NavLink, Text, UnstyledButton } from '@mantine/core';
import { useDisclosure } from '@mantine/hooks';
import { IconChevronDown, IconLogout } from '@tabler/icons-react';
import { Link, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';

interface NavItem {
  label: string;
  to: string;
  permission?: string;
}
// must be organized in the same concern of eg.. management users, management inventories, management projects, settings, 
const navItems: NavItem[] = [
  { label: 'Dashboard', to: '/dashboard' },
  { label: 'Projects', to: '/projects', permission: 'project.view' },
  { label: 'Users', to: '/settings/users', permission: 'user.view' },
  { label: 'Client Users', to: '/clients/users', permission: 'client_user.view' },
  { label: 'Client Companies', to: '/clients/companies', permission: 'client_company.view' },
  { label: 'Company Profile', to: '/settings/company', permission: 'owner_company.view' },
  { label: 'Roles & Permissions', to: '/settings/roles', permission: 'roles.view' },
  { label: 'Labels', to: '/settings/labels', permission: 'labels.view' },
];

export function MainLayout() {
  const [opened, { toggle }] = useDisclosure();
  const { user, can, logout } = useAuth();
  const location = useLocation();

  return (
    <AppShell
      header={{ height: 60 }}
      navbar={{ width: 240, breakpoint: 'sm', collapsed: { mobile: !opened } }}
      padding="md"
    >
      <AppShell.Header>
        <Group h="100%" px="md" justify="space-between">
          <Group>
            <Burger opened={opened} onClick={toggle} hiddenFrom="sm" size="sm" />
            <Text fw={700}>ConstructionPM</Text>
          </Group>
          <Menu position="bottom-end" withArrow>
            <Menu.Target>
              <UnstyledButton>
                <Group gap="xs">
                  <Avatar radius="xl" size="sm">
                    {user?.name.slice(0, 1)}
                  </Avatar>
                  <Text size="sm">{user?.name}</Text>
                  <IconChevronDown size={14} />
                </Group>
              </UnstyledButton>
            </Menu.Target>
            <Menu.Dropdown>
              <Menu.Item leftSection={<IconLogout size={14} />} onClick={() => logout()}>
                Log out
              </Menu.Item>
            </Menu.Dropdown>
          </Menu>
        </Group>
      </AppShell.Header>

      <AppShell.Navbar p="md">
        {navItems
          .filter((item) => !item.permission || can(item.permission))
          .map((item) => (
            <NavLink
              key={item.to}
              component={Link}
              to={item.to}
              label={item.label}
              active={location.pathname === item.to}
            />
          ))}
      </AppShell.Navbar>

      <AppShell.Main>
        <Outlet />
      </AppShell.Main>
    </AppShell>
  );
}
