import { AppShell, Avatar, Burger, Group, Menu, NavLink, ScrollArea, Stack, Text, Tooltip, UnstyledButton } from '@mantine/core';
import { useDisclosure } from '@mantine/hooks';
import { useState, useEffect } from 'react';
import { LayoutChromeContext } from './LayoutChromeContext';
import {
  IconActivity,
  IconBuilding,
  IconBuildingSkyscraper,
  IconBuildingWarehouse,
  IconCalendar,
  IconChartBar,
  IconChevronDown,
  IconCircleDot,
  IconCompass,
  IconFolder,
  IconForklift,
  IconHelpCircle,
  IconLayoutGrid,
  IconLayoutSidebarLeftCollapse,
  IconLayoutSidebarLeftExpand,
  IconListDetails,
  IconLogout,
  IconPresentationAnalytics,
  IconReportAnalytics,
  IconSettings,
  IconShieldLock,
  IconTag,
  IconUser,
  IconUserCheck,
  IconUserCircle,
  IconUsers,
} from '@tabler/icons-react';
import { Link, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';

interface NavItem {
  label: string;
  to?: string;
  permission?: string;
  icon?: typeof IconUsers;
  children?: NavItem[];
}

const navItems: NavItem[] = [
  {
    label: 'Analytics',
    icon: IconChartBar,
    children: [
      { label: 'Dashboard', to: '/dashboard', permission: 'project.view', icon: IconListDetails },
      { label: 'EVM Methodology', to: '#', permission: '#', icon:  IconPresentationAnalytics },
      { label: 'SMART Methodology', to: '#', permission: '#', icon:  IconReportAnalytics },
      { label: 'Log Activity', to: '#', permission: '#', icon:  IconActivity},
    ],
  },
  {
    label: 'Workspaces',
    icon: IconLayoutGrid,
    children: [
      { label: 'Projects', to: '/projects', permission: 'project.view', icon: IconListDetails },
      { label: 'Assigned to me', to: '#', permission: '#', icon: IconUserCheck },
      { label: 'Calendar', to: '#', permission: '#', icon: IconCalendar },
    ],
  },
  {
    label: 'Resources',
    icon: IconBuildingWarehouse,
    children: [
      { label: 'Inventories', to: '#', permission: '#', icon: IconFolder },
      { label: 'Suppliers', to: '#', permission: '#', icon: IconForklift },
      { label: 'Categories', to: '#', permission: '#', icon: IconTag },
    ],
  },
  {
    label: 'Users',
    icon: IconUsers,
    children: [
      { label: 'Team', to: '/settings/users', permission: 'user.view', icon: IconUser },
      { label: 'Client Users', to: '/clients/users', permission: 'client_user.view', icon: IconUserCircle },
      {
        label: 'Client Companies',
        to: '/clients/companies',
        permission: 'client_company.view',
        icon: IconBuildingSkyscraper,
      },
    ],
  },
  {
    label: 'Reports',
    icon: IconReportAnalytics,
    children: [
      { label: '#', to: '#', permission: '#', icon: IconCircleDot },
      { label: '#', to: '#', permission: '#', icon: IconCircleDot },
      { label: '#', to: '#', permission: '#', icon: IconCircleDot },
    ],
  },
  {
    label: 'Settings',
    icon: IconSettings,
    children: [
      { label: 'Company Profile', to: '/settings/company', permission: 'owner_company.view', icon: IconBuilding },
      { label: 'Roles & Permissions', to: '/settings/roles', permission: 'roles.view', icon: IconShieldLock },
      { label: 'Tags', to: '/settings/labels', permission: 'labels.view', icon: IconTag },
    ],
  },
];

function NavItemLink({
  item,
  can,
  pathname,
  collapsed,
}: {
  item: NavItem;
  can: (permission: string) => boolean;
  pathname: string;
  collapsed: boolean;
}) {
  if (item.children) {
    const visibleChildren = item.children.filter((child) => !child.permission || can(child.permission));
    if (visibleChildren.length === 0) {
      return null;
    }
    const hasActiveChild = visibleChildren.some((child) => child.to === pathname);

    if (collapsed) {
      return (
        <Tooltip label={item.label} position="right" withArrow>
          <NavLink
            label={item.label}
            leftSection={item.icon ? <item.icon size={16} stroke={1.5} /> : undefined}
            active={hasActiveChild}
          />
        </Tooltip>
      );
    }

    return (
      <NavLink
        label={item.label}
        leftSection={item.icon ? <item.icon size={16} stroke={1.5} /> : undefined}
        defaultOpened={hasActiveChild}
        childrenOffset={20}
      >
        {visibleChildren.map((child) => (
          <NavLink
            key={child.to}
            component={Link}
            to={child.to!}
            label={child.label}
            leftSection={child.icon ? <child.icon size={16} stroke={1.5} /> : undefined}
            active={pathname === child.to}
          />
        ))}
      </NavLink>
    );
  }
  if (item.permission && !can(item.permission)) {
    return null;
  }

  const linkContent = (
    <NavLink
      component={Link}
      to={item.to!}
      label={collapsed ? undefined : item.label}
      leftSection={item.icon ? <item.icon size={16} stroke={1.5} /> : undefined}
      active={pathname === item.to}
    />
  );

  return collapsed ? (
    <Tooltip label={item.label} position="right" withArrow>
      {linkContent}
    </Tooltip>
  ) : (
    linkContent
  );
}

// A single footer row: icon + label, matching the reference screenshot style
function FooterItem({
  icon: Icon,
  label,
  collapsed,
  onClick,
}: {
  icon: typeof IconHelpCircle;
  label: string;
  collapsed: boolean;
  onClick?: () => void;
}) {
  const content = (
    <UnstyledButton
      onClick={onClick}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 8,
        padding: '6px 8px',
        borderRadius: 4,
        width: '100%',
        justifyContent: collapsed ? 'center' : 'flex-start',
      }}
    >
      <Icon size={16} stroke={1.5} />
      {!collapsed && (
        <Text size="sm">
          {label}
        </Text>
      )}
    </UnstyledButton>
  );

  return collapsed ? (
    <Tooltip label={label} position="right" withArrow>
      {content}
    </Tooltip>
  ) : (
    content
  );
}

export function MainLayout() {
  const [mobileOpened, { toggle: toggleMobile }] = useDisclosure();
  const [desktopCollapsed, { toggle: toggleDesktop }] = useDisclosure(false);
  const { user, can, logout } = useAuth();
  const [fullscreen, setFullscreen] = useState(false);
  const location = useLocation();

  useEffect(() => {
    setFullscreen(false);
  }, [location.pathname]);

  const handleStartTour = () => {
    // Hook this up to your tour library, e.g. driver.js, react-joyride, intro.js
    // Example: tourInstance.drive();
    window.dispatchEvent(new CustomEvent('start-tour-guide'));
  };

  return (
    <AppShell
      header={{ height: 60, collapsed: fullscreen }}
      navbar={{
        width: desktopCollapsed ? 80 : 240,
        breakpoint: 'sm',
        collapsed: { mobile: !mobileOpened, desktop: fullscreen && !mobileOpened },
      }}
      padding={fullscreen ? 0 : 'md'}
      transitionDuration={200}
    >
      <AppShell.Header>
        <Group h="100%" px="md" justify="space-between">
          <Group>
            <Burger opened={mobileOpened} onClick={toggleMobile} hiddenFrom="sm" size="sm" />
            <Text fw={700} size="lg">
              Construction Project Manager
            </Text>
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

      <AppShell.Navbar p={0} style={{ display: 'flex', flexDirection: 'column' }}>
        <ScrollArea style={{ flex: 1 }} p="md" scrollbarSize={6}>
          {navItems.map((item) => (
            <NavItemLink
              key={item.label}
              item={item}
              can={can}
              pathname={location.pathname}
              collapsed={desktopCollapsed}
            />
          ))}
        </ScrollArea>
        {/* // feat: need fix */}
        <Stack gap={2} p="xs">
          <FooterItem
            icon={IconHelpCircle}
            label="Frequently Asked Questions"
            collapsed={desktopCollapsed}
          />
          <FooterItem
            icon={IconCompass}
            label="Tour guide"
            collapsed={desktopCollapsed}
            onClick={handleStartTour}
          />
          <FooterItem
            icon={desktopCollapsed ? IconLayoutSidebarLeftExpand : IconLayoutSidebarLeftCollapse}
            label={desktopCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
            collapsed={desktopCollapsed}
            onClick={toggleDesktop}
          />
        </Stack>
      </AppShell.Navbar>

      <AppShell.Main>
        <LayoutChromeContext.Provider value={{ fullscreen, setFullscreen, onOpenNav: toggleMobile }}>
          <Outlet />
        </LayoutChromeContext.Provider>
      </AppShell.Main>

    </AppShell>
  );
}