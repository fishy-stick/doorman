export type DdnsProviderType = '' | 'dnspod'

export type DNSPodConfigForm = {
  domain: string
  record: string
  id: string
  token: string
}

export type NetworkFormValues = {
  name: string
  token: string
  ddnsEnabled: boolean
  ddnsType: DdnsProviderType
  dnspod: DNSPodConfigForm
}

export type NetworkConfigMode =
  | { kind: 'structured-dnspod'; values: DNSPodConfigForm }
  | { kind: 'raw-json'; raw: string; reason: string }

export const emptyDNSPodConfig: DNSPodConfigForm = {
  domain: '',
  record: '',
  id: '',
  token: '',
}

export function emptyNetworkFormValues(): NetworkFormValues {
  return {
    name: '',
    token: '',
    ddnsEnabled: false,
    ddnsType: '',
    dnspod: { ...emptyDNSPodConfig },
  }
}

export function serializeDNSPodConfig(config: DNSPodConfigForm): string {
  return JSON.stringify({
    domain: config.domain.trim(),
    record: config.record.trim(),
    id: config.id.trim(),
    token: config.token.trim(),
  })
}

export function parseDNSPodConfig(raw: string): NetworkConfigMode {
  let parsed: unknown

  try {
    parsed = JSON.parse(raw || '{}')
  } catch {
    return {
      kind: 'raw-json',
      raw,
      reason: 'Current DDNS config is not valid JSON.',
    }
  }

  if (!isRecord(parsed)) {
    return {
      kind: 'raw-json',
      raw,
      reason: 'Current DDNS config cannot be mapped to the DNSPod form because it is not a JSON object.',
    }
  }

  const values = toDNSPodConfig(parsed)
  if (!values) {
    return {
      kind: 'raw-json',
      raw,
      reason: 'Current DDNS config cannot be mapped to the DNSPod form. Fix the JSON fields and save again.',
    }
  }

  return {
    kind: 'structured-dnspod',
    values,
  }
}

export function isSupportedProvider(value: string): value is DdnsProviderType {
  return value === '' || value === 'dnspod'
}

export function validateRawJSONConfig(raw: string): string | null {
  let parsed: unknown

  try {
    parsed = JSON.parse(raw || '{}')
  } catch {
    return 'DDNS config must be valid JSON.'
  }

  if (!isRecord(parsed)) {
    return 'DDNS config must be a JSON object.'
  }

  return null
}

function toDNSPodConfig(value: Record<string, unknown>): DNSPodConfigForm | null {
  const domain = readString(value, 'domain')
  const record = readString(value, 'record')
  const id = readString(value, 'id')
  const token = readString(value, 'token')

  if (domain === null || record === null || id === null || token === null) {
    return null
  }

  return {
    domain,
    record,
    id,
    token,
  }
}

function readString(value: Record<string, unknown>, key: string): string | null {
  const field = value[key]
  if (typeof field !== 'string') {
    return null
  }

  return field
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === 'object' && !Array.isArray(value)
}
