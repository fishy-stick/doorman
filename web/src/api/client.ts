import { ApiError } from '../utils/errors'

type UnauthorizedHandler = (() => void) | null

type ApiRequestOptions = {
  method?: string
  body?: unknown
  headers?: HeadersInit
  skipUnauthorized?: boolean
}

let unauthorizedHandler: UnauthorizedHandler = null

export function setUnauthorizedHandler(handler: UnauthorizedHandler): void {
  unauthorizedHandler = handler
}

export async function apiRequest<T>(path: string, options: ApiRequestOptions = {}): Promise<T> {
  const headers = new Headers(options.headers)
  const init: RequestInit = {
    method: options.method ?? 'GET',
    credentials: 'include',
    headers,
  }

  if (options.body !== undefined) {
    headers.set('Content-Type', 'application/json')
    init.body = JSON.stringify(options.body)
  }

  const response = await fetch(path, init)
  const payload = await parsePayload(response)

  if (!response.ok) {
    const message = extractErrorMessage(payload, response.statusText || 'Request failed')
    const error = new ApiError(response.status, message, payload)

    if (response.status === 401 && !options.skipUnauthorized) {
      unauthorizedHandler?.()
    }

    throw error
  }

  return payload as T
}

async function parsePayload(response: Response): Promise<unknown> {
  if (response.status === 204) {
    return undefined
  }

  const contentType = response.headers.get('content-type') ?? ''
  if (contentType.includes('application/json')) {
    return response.json().catch(() => undefined)
  }

  const text = await response.text()
  return text === '' ? undefined : text
}

function extractErrorMessage(payload: unknown, fallback: string): string {
  if (payload && typeof payload === 'object' && 'error' in payload) {
    const value = (payload as { error?: unknown }).error
    if (typeof value === 'string' && value.trim() !== '') {
      return value
    }
  }

  if (typeof payload === 'string' && payload.trim() !== '') {
    return payload
  }

  return fallback
}
