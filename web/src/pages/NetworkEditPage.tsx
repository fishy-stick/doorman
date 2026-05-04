import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { getNetwork, updateNetwork, type NetworkDetail } from '../api/networks'
import { NetworkForm } from '../components/forms/NetworkForm'
import { ErrorState } from '../components/feedback/ErrorState'
import { LoadingState } from '../components/feedback/LoadingState'
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
  compatibilityReason?: string
  networkName: string
}

export function NetworkEditPage() {
  const navigate = useNavigate()
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

        setMessage(errorMessage(error, 'Unable to load the network.'))
      })
      .finally(() => {
        if (active) {
          setLoading(false)
        }
      })

    return () => {
      active = false
    }
  }, [networkId])

  if (!networkId) {
    return <ErrorState title="Network not found" message="The requested network does not exist." />
  }

  if (loading) {
    return <LoadingState label="Loading network" />
  }

  if (notFound) {
    return <ErrorState title="Network not found" message="The requested network does not exist." />
  }

  if (message || !formState || !networkId) {
    return <ErrorState message={message || 'Unable to load the network.'} actionLabel="Retry" onAction={() => window.location.reload()} />
  }

  return (
    <section>
      <div className="page-header">
        <div>
          <h1>Edit Network</h1>
          <p>Update {formState.networkName} without changing its history.</p>
        </div>
      </div>
      <NetworkForm
        initialValues={formState.initialValues}
        initialRawConfig={formState.initialRawConfig}
        compatibilityReason={formState.compatibilityReason}
        submitLabel="Save Changes"
        submittingLabel="Saving..."
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
      compatibilityReason: `Current provider "${network.ddns_type}" is not supported by this frontend. Choose a supported provider before saving.`,
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
      compatibilityReason: parsed.reason,
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
