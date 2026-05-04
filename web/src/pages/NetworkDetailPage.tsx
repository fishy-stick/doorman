import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { deleteNetwork, getNetwork, regenerateNetworkToken, type NetworkDetail } from '../api/networks'
import { Badge } from '../components/data-display/Badge'
import { DdnsStatusBadge } from '../components/data-display/DdnsStatusBadge'
import { FieldList } from '../components/data-display/FieldList'
import { ErrorState } from '../components/feedback/ErrorState'
import { LoadingState } from '../components/feedback/LoadingState'
import { formatDate } from '../utils/date'
import { errorMessage, isApiError } from '../utils/errors'

export function NetworkDetailPage() {
  const navigate = useNavigate()
  const { networkId } = useParams()
  const [loading, setLoading] = useState(true)
  const [message, setMessage] = useState('')
  const [notFound, setNotFound] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [regeneratingToken, setRegeneratingToken] = useState(false)
  const [copyMessage, setCopyMessage] = useState('')
  const [network, setNetwork] = useState<NetworkDetail | null>(null)

  useEffect(() => {
    if (!networkId) {
      return
    }

    let active = true

    getNetwork(networkId)
      .then((result) => {
        if (active) {
          setNetwork(result)
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

  async function handleDelete() {
    if (!networkId || !network) {
      return
    }

    const confirmed = window.confirm(`Delete network "${network.name}"? This also removes its knock history.`)
    if (!confirmed) {
      return
    }

    setDeleting(true)
    setMessage('')

    try {
      await deleteNetwork(networkId)
      navigate('/admin/networks', { replace: true })
    } catch (error) {
      setMessage(errorMessage(error, 'Unable to delete the network.'))
      setDeleting(false)
    }
  }

  async function handleCopy(value: string, label: string) {
    try {
      await navigator.clipboard.writeText(value)
      setCopyMessage(`${label} copied.`)
    } catch {
      setCopyMessage(`Unable to copy ${label.toLowerCase()}.`)
    }
  }

  async function handleRegenerateToken() {
    if (!networkId || !network) {
      return
    }

    const confirmed = window.confirm(
      `Regenerate token for "${network.name}"? Existing clients using the current token will stop working until their commands are updated.`,
    )
    if (!confirmed) {
      return
    }

    setRegeneratingToken(true)
    setMessage('')
    setCopyMessage('')

    try {
      const updated = await regenerateNetworkToken(networkId)
      setNetwork(updated)
      setCopyMessage('Token regenerated. Update any clients that use the old command.')
    } catch (error) {
      setMessage(errorMessage(error, 'Unable to regenerate token.'))
    } finally {
      setRegeneratingToken(false)
    }
  }

  if (!networkId) {
    return <ErrorState title="Network not found" message="The requested network does not exist." />
  }

  if (loading) {
    return <LoadingState label="Loading network detail" />
  }

  if (notFound) {
    return <ErrorState title="Network not found" message="The requested network does not exist." />
  }

  if (message && !network) {
    return <ErrorState message={message} actionLabel="Retry" onAction={() => window.location.reload()} />
  }

  if (!network) {
    return <ErrorState message="Unable to load the network." />
  }

  return (
    <section>
      <div className="page-header">
        <div>
          <h1>{network.name}</h1>
          <p>Inspect the current IP snapshot, DDNS behavior, and generated client commands.</p>
        </div>
        <div className="page-actions">
          <Link className="button button-secondary" to={`/admin/networks/${network.id}/edit`}>
            Edit
          </Link>
          <Link className="button button-secondary" to={`/admin/networks/${network.id}/history`}>
            History
          </Link>
          <button className="button button-secondary button-danger" type="button" disabled={deleting} onClick={handleDelete}>
            {deleting ? 'Deleting...' : 'Delete'}
          </button>
        </div>
      </div>

      {message ? <div className="panel-stack"><div className="form-message form-message-error">{message}</div></div> : null}
      {copyMessage ? <div className="panel-stack"><div className="form-message form-message-success">{copyMessage}</div></div> : null}

      <section className="page-panel">
        <FieldList
          items={[
            { label: 'Name', value: network.name },
            {
              label: 'Token',
              value: (
                <div className="inline-actions">
                  <span className="mono">{network.token}</span>
                  <button className="button button-secondary" type="button" onClick={() => handleCopy(network.token, 'token')}>
                    Copy
                  </button>
                  <button
                    className="button button-secondary button-danger"
                    type="button"
                    disabled={regeneratingToken}
                    onClick={handleRegenerateToken}
                  >
                    {regeneratingToken ? 'Regenerating...' : 'Regenerate'}
                  </button>
                </div>
              ),
            },
            {
              label: 'DDNS',
              value: <Badge tone={network.ddns_enabled ? 'success' : 'neutral'}>{network.ddns_enabled ? 'Enabled' : 'Disabled'}</Badge>,
            },
            { label: 'Provider', value: network.ddns_type || 'None' },
            { label: 'Current IP', value: network.current_ip ?? 'Unknown', mono: true },
            { label: 'Previous IP', value: network.previous_ip ?? 'None', mono: true },
            { label: 'Last Knock', value: formatDate(network.last_knock) },
            { label: 'Latest DDNS Status', value: <DdnsStatusBadge status={network.ddns_status} /> },
          ]}
        />
      </section>

      <section className="page-panel command-panel">
        <div className="section-heading">
          <div>
            <h2>Client Commands</h2>
            <p>Use these as a starting point, then replace `your-server:8080` with the real host and port.</p>
          </div>
        </div>

        <div className="form-message form-message-info">
          <strong>Host placeholder</strong>
          <p>The generated commands still use `your-server:8080`. Update that part before deploying them.</p>
        </div>

        <div className="command-grid">
          <article className="command-card">
            <div className="command-card-header">
              <h3>curl</h3>
              <button className="button button-secondary" type="button" onClick={() => handleCopy(network.commands.curl, 'curl command')}>
                Copy
              </button>
            </div>
            <pre className="code-block">
              <code>{network.commands.curl}</code>
            </pre>
          </article>

          <article className="command-card">
            <div className="command-card-header">
              <h3>crontab</h3>
              <button className="button button-secondary" type="button" onClick={() => handleCopy(network.commands.crontab, 'crontab command')}>
                Copy
              </button>
            </div>
            <pre className="code-block">
              <code>{network.commands.crontab}</code>
            </pre>
          </article>
        </div>
      </section>
    </section>
  )
}
