import { apiRequest } from './client'

export type NetworkSummary = {
  id: number
  name: string
  ddns_enabled: boolean
  ddns_type: string
  current_ip: string | null
  previous_ip: string | null
  last_knock: string | null
  ddns_status: string | null
}

export type NetworkDetail = {
  id: number
  name: string
  token: string
  ddns_enabled: boolean
  ddns_type: string
  ddns_config: string
  current_ip: string | null
  previous_ip: string | null
  last_knock: string | null
  ddns_status: string | null
  commands: {
    curl: string
    crontab: string
  }
}

export type NetworkPayload = {
  name: string
  token: string
  ddns_enabled: boolean
  ddns_type: string
  ddns_config: string
}

export type KnockRecord = {
  id: number
  network_id: number
  ip: string
  previous_ip: string | null
  ip_changed: boolean
  user_agent: string
  ddns_status: string
  ddns_error: string
  created_at: string
}

export type KnockListResponse = {
  total: number
  page: number
  size: number
  records: KnockRecord[]
}

export async function listNetworks(): Promise<NetworkSummary[]> {
  const response = await apiRequest<{ networks: NetworkSummary[] }>('/admin/api/networks')
  return response.networks
}

export function getNetwork(id: string | number): Promise<NetworkDetail> {
  return apiRequest<NetworkDetail>(`/admin/api/networks/${id}`)
}

export function createNetwork(payload: NetworkPayload): Promise<NetworkDetail> {
  return apiRequest<NetworkDetail>('/admin/api/networks', {
    method: 'POST',
    body: payload,
  })
}

export function updateNetwork(id: string | number, payload: NetworkPayload): Promise<NetworkDetail> {
  return apiRequest<NetworkDetail>(`/admin/api/networks/${id}`, {
    method: 'PUT',
    body: payload,
  })
}

export function deleteNetwork(id: string | number): Promise<{ message: string }> {
  return apiRequest<{ message: string }>(`/admin/api/networks/${id}`, {
    method: 'DELETE',
  })
}

export function listKnocks(id: string | number, page: number, size: number): Promise<KnockListResponse> {
  const params = new URLSearchParams({
    page: String(page),
    size: String(size),
  })

  return apiRequest<KnockListResponse>(`/admin/api/networks/${id}/knocks?${params.toString()}`)
}
