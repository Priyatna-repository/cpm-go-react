import { useEffect, useRef, useState, type ChangeEvent, type FormEvent } from 'react';
import {
  Alert, Avatar, Box, Button, Group, Paper, Select, Stack, TextInput, Title,
} from '@mantine/core';
import { IconAlertTriangle } from '@tabler/icons-react';
import * as ownerCompanyApi from '../../api/ownerCompany';
import type { OwnerCompany } from '../../api/ownerCompany';
import * as lookupsApi from '../../api/lookups';
import type { Country, Currency } from '../../api/lookups';
import { useAuth } from '../../auth/AuthContext';
import { errorMessage } from '../../utils/errors';

export function CompanyPage() {
  const { can } = useAuth();
  const canEdit = can('owner_company.edit');

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  const [countries, setCountries] = useState<Country[]>([]);
  const [currencies, setCurrencies] = useState<Currency[]>([]);
  const [logoUrl, setLogoUrl] = useState<string | undefined>(undefined);
  const [logoFile, setLogoFile] = useState<File | null>(null);
  const [logoPreview, setLogoPreview] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [name, setName] = useState('');
  const [address, setAddress] = useState('');
  const [postalCode, setPostalCode] = useState('');
  const [city, setCity] = useState('');
  const [countryId, setCountryId] = useState<string | null>(null);
  const [currencyId, setCurrencyId] = useState<string | null>(null);
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('');
  const [web, setWeb] = useState('');

  function applyCompany(company: OwnerCompany) {
    setName(company.name);
    setAddress(company.address ?? '');
    setPostalCode(company.postal_code ?? '');
    setCity(company.city ?? '');
    setCountryId(company.country_id ? String(company.country_id) : null);
    setCurrencyId(company.currency_id ? String(company.currency_id) : null);
    setEmail(company.email ?? '');
    setPhone(company.phone ?? '');
    setWeb(company.web ?? '');
    setLogoUrl(company.logo ? `${import.meta.env.VITE_API_BASE_URL}${company.logo}` : undefined);
  }

  useEffect(() => {
    (async () => {
      try {
        const [company, countryList, currencyList] = await Promise.all([
          ownerCompanyApi.getOwnerCompany(),
          lookupsApi.listCountries(),
          lookupsApi.listCurrencies(),
        ]);
        applyCompany(company);
        setCountries(countryList);
        setCurrencies(currencyList);
      } catch (err) {
        setError(errorMessage(err, 'Failed to load company profile.'));
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  function handleLogoChange(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setLogoPreview((prev) => {
      if (prev) URL.revokeObjectURL(prev);
      return URL.createObjectURL(file);
    });
    setLogoFile(file);
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);
    setSaved(false);
    try {
      const updated = await ownerCompanyApi.updateOwnerCompany({
        name,
        address: address || undefined,
        postal_code: postalCode || undefined,
        city: city || undefined,
        country_id: countryId ? Number(countryId) : undefined,
        currency_id: currencyId ? Number(currencyId) : undefined,
        email: email || undefined,
        phone: phone || undefined,
        web: web || undefined,
        logo: logoFile,
      });
      applyCompany(updated);
      setLogoFile(null);
      setLogoPreview(null);
      setSaved(true);
    } catch (err) {
      setError(errorMessage(err, 'Failed to save changes.'));
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return null;
  }

  return (
    <Stack p="xl" maw={560}>
      <Title order={2}>Company Profile</Title>

      {error && (
        <Alert color="red" icon={<IconAlertTriangle size={18} />}>
          {error}
        </Alert>
      )}
      {saved && <Alert color="green">Saved.</Alert>}

      <Paper radius="md" p="lg" withBorder>
        <form onSubmit={handleSubmit}>
          <Group align="flex-start">
            <Box>
              <Avatar
                src={logoPreview ?? logoUrl}
                size={100}
                radius="md"
                style={{ cursor: canEdit ? 'pointer' : 'default' }}
                onClick={() => canEdit && fileInputRef.current?.click()}
              >
                {name.slice(0, 1)}
              </Avatar>
              {canEdit && (
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/png,image/jpeg,image/gif"
                  hidden
                  onChange={handleLogoChange}
                />
              )}
            </Box>
            <Stack gap="xs" style={{ flex: 1 }}>
              <TextInput
                label="Name"
                required
                disabled={!canEdit}
                value={name}
                onChange={(e) => setName(e.currentTarget.value)}
              />
              <Select
                label="Currency"
                disabled={!canEdit}
                searchable
                data={currencies.map((c) => ({ value: String(c.id), label: `${c.name} (${c.symbol})` }))}
                value={currencyId}
                onChange={setCurrencyId}
              />
            </Stack>
          </Group>

          <TextInput mt="md" label="Address" disabled={!canEdit} value={address} onChange={(e) => setAddress(e.currentTarget.value)} />
          <Group grow mt="md">
            <TextInput label="City" disabled={!canEdit} value={city} onChange={(e) => setCity(e.currentTarget.value)} />
            <TextInput label="Postal code" disabled={!canEdit} value={postalCode} onChange={(e) => setPostalCode(e.currentTarget.value)} />
          </Group>
          <Select
            mt="md"
            label="Country"
            disabled={!canEdit}
            searchable
            data={countries.map((c) => ({ value: String(c.id), label: c.name }))}
            value={countryId}
            onChange={setCountryId}
          />
          <TextInput mt="md" label="Email" disabled={!canEdit} value={email} onChange={(e) => setEmail(e.currentTarget.value)} />
          <Group grow mt="md">
            <TextInput label="Phone" disabled={!canEdit} value={phone} onChange={(e) => setPhone(e.currentTarget.value)} />
            <TextInput label="Website" disabled={!canEdit} value={web} onChange={(e) => setWeb(e.currentTarget.value)} />
          </Group>

          {canEdit && (
            <Button type="submit" mt="xl" loading={saving}>
              Save
            </Button>
          )}
        </form>
      </Paper>
    </Stack>
  );
}
