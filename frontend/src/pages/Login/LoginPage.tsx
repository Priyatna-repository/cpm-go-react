import { useCallback, useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import {
  Alert,
  Button,
  Checkbox,
  Divider,
  Group,
  Paper,
  PasswordInput,
  Text,
  TextInput,
  Title,
} from '@mantine/core';
import { IconAlertTriangle, IconLock, IconMail } from '@tabler/icons-react';
import { GuestLayout } from '../../layouts/GuestLayout';
import { GoogleButton } from './GoogleButton';
import { useAuth } from '../../auth/AuthContext';

export function LoginPage() {
  const navigate = useNavigate();
  const { login, loginWithGoogle } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [remember, setRemember] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function errorMessage(err: unknown, fallback: string): string {
    if (axios.isAxiosError(err) && typeof err.response?.data?.error === 'string') {
      return err.response.data.error;
    }
    return fallback;
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      await login(email, password);
      navigate('/dashboard');
    } catch (err) {
      setError(errorMessage(err, 'Login failed'));
    } finally {
      setSubmitting(false);
    }
  }

  const handleGoogleCredential = useCallback(
    async (credential: string) => {
      setError(null);
      try {
        await loginWithGoogle(credential);
        navigate('/dashboard');
      } catch (err) {
        setError(errorMessage(err, 'Google sign-in failed.'));
      }
    },
    [loginWithGoogle, navigate],
  );

  return (
    <GuestLayout>
      <Paper radius="md" p="lg" withBorder w={380}>
        <Title order={2} ta="center">
          Log in to Your ConstPM Account
        </Title>
        <Text c="dimmed" size="sm" ta="center" mt={5}>
          Access your projects and tasks securely.
        </Text>

        {error && (
          <Alert
            radius="md"
            title="Login failed"
            icon={<IconAlertTriangle size={18} />}
            color="red"
            mt="md"
          >
            {error}
          </Alert>
        )}

        <form onSubmit={handleSubmit}>
          <TextInput
            label="Email"
            placeholder="Your email"
            required
            leftSection={<IconMail size={18} stroke={1.5} />}
            value={email}
            onChange={(e) => setEmail(e.currentTarget.value)}
            size="md"
            radius="md"
            mt="md"
          />
          <PasswordInput
            label="Password"
            placeholder="Your password"
            required
            leftSection={<IconLock size={18} stroke={1.5} />}
            value={password}
            onChange={(e) => setPassword(e.currentTarget.value)}
            size="md"
            radius="md"
            mt="md"
          />
          <Group justify="space-between" mt="lg">
            <Checkbox
              label="Remember me"
              checked={remember}
              onChange={(e) => setRemember(e.currentTarget.checked)}
            />
          </Group>
          <Button type="submit" fullWidth mt="xl" size="md" radius="md" loading={submitting}>
            Login
          </Button>
        </form>

        <Divider label="or" labelPosition="center" my="lg" />
        <GoogleButton onCredential={handleGoogleCredential} />
      </Paper>
    </GuestLayout>
  );
}
