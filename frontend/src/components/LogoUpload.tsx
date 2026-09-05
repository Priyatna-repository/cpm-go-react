import { useRef, useState, type ChangeEvent } from 'react';
import { Avatar, Box, Text } from '@mantine/core';

interface LogoUploadProps {
  currentUrl?: string;
  fallbackText: string;
  disabled?: boolean;
  onChange: (file: File) => void;
}

export function LogoUpload({ currentUrl, fallbackText, disabled, onChange }: LogoUploadProps) {
  const [preview, setPreview] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  function handleChange(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setPreview((prev) => {
      if (prev) URL.revokeObjectURL(prev);
      return URL.createObjectURL(file);
    });
    onChange(file);
  }

  return (
    <Box>
      <Avatar
        src={preview ?? currentUrl}
        size={100}
        radius="md"
        style={{ cursor: disabled ? 'default' : 'pointer' }}
        onClick={() => !disabled && fileInputRef.current?.click()}
      >
        {fallbackText.slice(0, 1)}
      </Avatar>
      {!disabled && (
        <>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/png,image/jpeg,image/gif"
            hidden
            onChange={handleChange}
          />
          <Text size="xs" c="dimmed" mt={4}>
            Click to change
          </Text>
        </>
      )}
    </Box>
  );
}
