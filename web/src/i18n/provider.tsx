import { useState, type ReactNode } from 'react'
import { I18nContext } from './context'
import { getLocale, setLocale, type Locale } from './locale'
import {
  translateForLocale,
  type TranslationKey,
  type TranslationValues,
} from './messages'

type I18nProviderProps = {
  children: ReactNode
}

export function I18nProvider({ children }: I18nProviderProps) {
  const [locale, setLocaleState] = useState<Locale>(() => getLocale())

  function handleSetLocale(nextLocale: Locale) {
    setLocale(nextLocale)
    setLocaleState(nextLocale)
  }

  function translate(key: TranslationKey, values?: TranslationValues): string {
    return translateForLocale(locale, key, values)
  }

  return (
    <I18nContext.Provider
      value={{
        locale,
        setLocale: handleSetLocale,
        t: translate,
      }}
    >
      {children}
    </I18nContext.Provider>
  )
}
