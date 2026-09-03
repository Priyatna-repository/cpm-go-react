import { Button, Stack, Text, Title } from '@mantine/core';
import { useAuth } from '../../auth/AuthContext';

export function DashboardPage() {
  const { user, logout } = useAuth();

  return (
    <Stack p="xl">
      <Title order={2}>Dashboard</Title>
      <Text>
        Logged in as {user?.name} ({user?.role})
      </Text>
      <Button w={160} onClick={() => logout()}>
        Log out
      </Button>
    </Stack>
  );
}
