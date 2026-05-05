import type { TranslationKey } from '../i18n/messages'

export type DdnsProviderType = '' | 'dnspod'

export type DNSPodConfigForm = {
  domain: string
  record: string
  id: string
  token: string
}

export type NetworkFormValues = {
  name: string
  ddnsEnabled: boolean
  ddnsType: DdnsProviderType
  dnspod: DNSPodConfigForm
}

export type NetworkConfigMode =
  | { kind: 'structured-dnspod'; values: DNSPodConfigForm }
  | { kind: 'raw-json'; raw: string; reasonKey: TranslationKey }

export const emptyDNSPodConfig: DNSPodConfigForm = {
  domain: '',
  record: '',
  id: '',
  token: '',
}

export function emptyNetworkFormValues(): NetworkFormValues {
  return {
    name: '',
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
      reasonKey: 'networkForm.compatibilityInvalidJson',
    }
  }

  if (!isRecord(parsed)) {
    return {
      kind: 'raw-json',
      raw,
      reasonKey: 'networkForm.compatibilityNotObject',
    }
  }

  const values = toDNSPodConfig(parsed)
  if (!values) {
    return {
      kind: 'raw-json',
      raw,
      reasonKey: 'networkForm.compatibilityUnsupportedShape',
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

export function validateRawJSONConfig(raw: string): TranslationKey | null {
  let parsed: unknown

  try {
    parsed = JSON.parse(raw || '{}')
  } catch {
    return 'networkForm.rawJsonInvalid'
  }

  if (!isRecord(parsed)) {
    return 'networkForm.rawJsonNotObject'
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
