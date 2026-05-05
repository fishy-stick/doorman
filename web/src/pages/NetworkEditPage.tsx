import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useI18n } from '../i18n'
import { getNetwork, updateNetwork, type NetworkDetail } from '../api/networks'
import { NetworkForm } from '../components/forms/NetworkForm'
import { ErrorState } from '../components/feedback/ErrorState'
import { LoadingState } from '../components/feedback/LoadingState'
import type { TranslationKey, TranslationValues } from '../i18n/messages'
import {
  emptyDNSPodConfig,
  isSupportedProvider,
  parseDNSPodConfig,
  type NetworkFormValues,
} from '../utils/ddns'
import { errorMessage, isApiError } from '../utils/errors'

type LoadedFormState = {
  initialValues: NetworkFormValues
  initialRawConfig?: string
  compatibilityReason?: {
    key: TranslationKey
    values?: TranslationValues
  }
  networkName: string
}

export function NetworkEditPage() {
  const navigate = useNavigate()
  const { t } = useI18n()
  const { networkId } = useParams()
  const [loading, setLoading] = useState(true)
  const [message, setMessage] = useState('')
  const [notFound, setNotFound] = useState(false)
  const [formState, setFormState] = useState<LoadedFormState | null>(null)

  useEffect(() => {
    if (!networkId) {
      return
    }

    let active = true

    getNetwork(networkId)
      .then((network) => {
        if (active) {
          setFormState(buildFormState(network))
        }
      })
      .catch((error) => {
        if (!active) {
          return
        }

        if (isApiError(error) && error.status === 404) {
          setNotFound(true)
          return
        }

        setMessage(errorMessage(error, t('networkEdit.unableLoad')))
      })
      .finally(() => {
        if (active) {
          setLoading(false)
        }
      })

    return () => {
      active = false
    }
  }, [networkId, t])

  if (!networkId) {
    return <ErrorState title={t('networkEdit.notFoundTitle')} message={t('networkEdit.notFoundMessage')} />
  }

  if (loading) {
    return <LoadingState label={t('networkEdit.loading')} />
  }

  if (notFound) {
    return <ErrorState title={t('networkEdit.notFoundTitle')} message={t('networkEdit.notFoundMessage')} />
  }

  if (message || !formState || !networkId) {
    return (
      <ErrorState
        message={message || t('networkEdit.unableLoad')}
        actionLabel={t('common.retry')}
        onAction={() => window.location.reload()}
      />
    )
  }

  return (
    <section>
      <div className="page-header">
        <div>
          <h1>{t('networkEdit.title')}</h1>
          <p>{t('networkEdit.subtitle', { name: formState.networkName })}</p>
        </div>
      </div>
      <NetworkForm
        initialValues={formState.initialValues}
        initialRawConfig={formState.initialRawConfig}
        compatibilityReason={
          formState.compatibilityReason
            ? t(formState.compatibilityReason.key, formState.compatibilityReason.values)
            : undefined
        }
        submitLabel={t('networkEdit.save')}
        submittingLabel={t('networkEdit.saving')}
        cancelTo={`/admin/networks/${networkId}`}
        onSubmit={async (payload) => {
          await updateNetwork(networkId, payload)
          navigate(`/admin/networks/${networkId}`, { replace: true })
        }}
      />
    </section>
  )
}

function buildFormState(network: NetworkDetail): LoadedFormState {
  const initialValues: NetworkFormValues = {
    name: network.name,
    ddnsEnabled: network.ddns_enabled,
    ddnsType: '',
    dnspod: { ...emptyDNSPodConfig },
  }

  if (!isSupportedProvider(network.ddns_type)) {
    return {
      initialValues,
      initialRawConfig: network.ddns_config,
      compatibilityReason: {
        key: 'networkEdit.unsupportedProviderReason',
        values: { provider: network.ddns_type },
      },
      networkName: network.name,
    }
  }

  initialValues.ddnsType = network.ddns_type

  if (network.ddns_type !== 'dnspod') {
    return {
      initialValues,
      networkName: network.name,
    }
  }

  const parsed = parseDNSPodConfig(network.ddns_config)
  if (parsed.kind === 'raw-json') {
    return {
      initialValues,
      initialRawConfig: parsed.raw,
      compatibilityReason: { key: parsed.reasonKey },
      networkName: network.name,
    }
  }

  return {
    initialValues: {
      ...initialValues,
      dnspod: parsed.values,
    },
    networkName: network.name,
  }
}
