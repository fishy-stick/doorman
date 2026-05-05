import type { Locale } from '../i18n/locale'

export function formatDate(value: string | null | undefined, locale: Locale, emptyLabel: string): string {
  if (!value) {
    return emptyLabel
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  return new Intl.DateTimeFormat(locale, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
}
