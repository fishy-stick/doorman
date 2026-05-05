import { createContext } from 'react'
import type { Locale } from './locale'
import type { TranslationKey, TranslationValues } from './messages'

export type Translator = (key: TranslationKey, values?: TranslationValues) => string

export type I18nContextValue = {
  locale: Locale
  setLocale: (locale: Locale) => void
  t: Translator
}

export const I18nContext = createContext<I18nContextValue | null>(null)
