import { apiRequest } from './client'

export type SessionStatus = {
  authenticated: true
}

type SessionOptions = {
  skipUnauthorized?: boolean
}

export function checkSession(options: SessionOptions = {}): Promise<SessionStatus> {
  return apiRequest<SessionStatus>('/admin/api/session', {
    skipUnauthorized: options.skipUnauthorized,
  })
}

export function login(password: string): Promise<{ message: string }> {
  return apiRequest<{ message: string }>('/admin/api/login', {
    method: 'POST',
    body: { password },
    skipUnauthorized: true,
  })
}

export function logout(): Promise<{ message: string }> {
  return apiRequest<{ message: string }>('/admin/api/logout', {
    method: 'POST',
    skipUnauthorized: true,
  })
}

export function changePassword(oldPassword: string, newPassword: string): Promise<{ message: string }> {
  return apiRequest<{ message: string }>('/admin/api/password', {
    method: 'PUT',
    body: {
      old_password: oldPassword,
      new_password: newPassword,
    },
  })
}
