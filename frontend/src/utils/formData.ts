export function appendIfPresent(form: FormData, key: string, value: string | number | undefined | null) {
  if (value !== undefined && value !== null && value !== '') {
    form.append(key, String(value));
  }
}
