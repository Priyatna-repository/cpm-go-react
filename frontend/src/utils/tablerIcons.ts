import * as TablerIcons from '@tabler/icons-react';

// Every exported icon component name, e.g. "IconFlag", "IconCheck" — used
// both to power the Label form's suggestion list and to safely resolve a
// stored (free-text, unvalidated) icon name back to a real component.
export const TABLER_ICON_NAMES = Object.keys(TablerIcons).filter((k) => k.startsWith('Icon'));

export function resolveIcon(name?: string) {
  if (!name) return null;
  const Icon = (TablerIcons as Record<string, any>)[name];
  return Icon != null ? Icon : null;
}
