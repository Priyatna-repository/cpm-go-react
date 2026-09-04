import { Button, Stack, Text, Title } from '@mantine/core';
import { Link } from 'react-router-dom';
import { useAuth } from '../../auth/AuthContext';

export function DashboardPage() {
  const { user, can, logout } = useAuth();

  return (
    <Stack p="xl">
      <Title order={2}>Dashboard</Title>
      <Text>
        Logged in as {user?.name} ({user?.role})
      </Text>
      {can('roles.view') && (
        <Button component={Link} to="/settings/roles" w={240} variant="light">
          Manage Roles & Permissions
        </Button>
      )}
      {can('owner_company.view') && (
        <Button component={Link} to="/settings/company" w={240} variant="light">
          Company Profile
        </Button>
      )}
      {can('client_company.view') && (
        <Button component={Link} to="/clients/companies" w={240} variant="light">
          Client Companies
        </Button>
      )}

      <Button w={160} onClick={() => logout()}>
        Log out
      </Button>
    </Stack>
  );
}
