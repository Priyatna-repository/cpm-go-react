import { Container } from '@mantine/core';
import type { ReactNode } from 'react';

export function GuestLayout({ children }: { children: ReactNode }) {
  return (
    <Container
      size="xs"
      h="100vh"
      style={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}
    >
      {children}
    </Container>
  );
}
