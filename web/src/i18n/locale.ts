export type Locale = 'zh-CN' | 'en'

export const localeStorageKey = 'doorman_locale'

function normalizeLocale(value: string | null | undefined): Locale | null {
  const normalized = value?.trim().toLowerCase() ?? ''

  if (normalized.startsWith('zh')) {
    return 'zh-CN'
  }

  if (normalized.startsWith('en')) {
    return 'en'
  }

  return null
}

function getStoredLocale(): Locale | null {
  if (typeof window === 'undefined') {
    return null
  }

  try {
    return normalizeLocale(window.localStorage.getItem(localeStorageKey))
  } catch {
    return null
  }
}

function getBrowserLocale(): Locale {
  if (typeof navigator === 'undefined') {
    return 'en'
  }

  const candidates = [...(navigator.languages ?? []), navigator.language]
  for (const candidate of candidates) {
    const locale = normalizeLocale(candidate)
    if (locale) {
      return locale
    }
  }

  return 'en'
}

function syncDocumentLanguage(locale: Locale): void {
  if (typeof document !== 'undefined') {
    document.documentElement.lang = locale
  }
}

function detectInitialLocale(): Locale {
  return getStoredLocale() ?? getBrowserLocale()
}

let currentLocale: Locale = detectInitialLocale()
syncDocumentLanguage(currentLocale)

export function getLocale(): Locale {
  return currentLocale
}

export function setLocale(locale: Locale): void {
  currentLocale = locale
  syncDocumentLanguage(locale)

  if (typeof window === 'undefined') {
    return
  }

  try {
    window.localStorage.setItem(localeStorageKey, locale)
  } catch {
    // Ignore storage errors and keep the in-memory locale.
  }
}
