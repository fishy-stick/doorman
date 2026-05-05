import { translateForLocale } from '../i18n/messages'
import { getLocale } from '../i18n/locale'

export class ApiError extends Error {
  status: number
  data: unknown

  constructor(status: number, message: string, data: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.data = data
  }
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError
}

export function isUnauthorized(error: unknown): boolean {
  return isApiError(error) && error.status === 401
}

export function errorMessage(error: unknown, fallback = translateForLocale(getLocale(), 'feedback.requestFailed')): string {
  if (error instanceof Error && error.message.trim() !== '') {
    return error.message
  }

  return fallback
}
