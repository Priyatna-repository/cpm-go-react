import { useEffect, useRef, useState } from 'react';
import { Text } from '@mantine/core';

declare global {
  interface Window {
    google?: {
      accounts: {
        id: {
          initialize: (config: {
            client_id: string;
            callback: (resp: { credential: string }) => void;
          }) => void;
          renderButton: (parent: HTMLElement, options: Record<string, unknown>) => void;
        };
      };
    };
  }
}

interface GoogleButtonProps {
  onCredential: (credential: string) => void;
}

export function GoogleButton({ onCredential }: GoogleButtonProps) {
  const divRef = useRef<HTMLDivElement>(null);
  const [loadFailed, setLoadFailed] = useState(false);

  useEffect(() => {
    const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID;
    if (!clientId || !divRef.current) return;

    function render() {
      if (!window.google || !divRef.current) return;
      window.google.accounts.id.initialize({
        client_id: clientId,
        callback: (resp) => onCredential(resp.credential),
      });
      window.google.accounts.id.renderButton(divRef.current, {
        theme: 'outline',
        size: 'large',
        width: 340,
      });
    }

    if (window.google) {
      render();
      return;
    }

    let attempts = 0;
    const maxAttempts = 50; // ~5s at 100ms intervals
    const interval = setInterval(() => {
      attempts += 1;
      if (window.google) {
        clearInterval(interval);
        render();
        return;
      }
      if (attempts >= maxAttempts) {
        clearInterval(interval);
        setLoadFailed(true);
      }
    }, 100);
    return () => clearInterval(interval);
  }, [onCredential]);

  if (loadFailed) {
    return (
      <Text c="dimmed" size="sm" ta="center">
        Google sign-in unavailable right now.
      </Text>
    );
  }

  return <div ref={divRef} />;
}
