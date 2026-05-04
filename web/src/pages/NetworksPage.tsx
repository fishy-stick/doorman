import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { deleteNetwork, listNetworks, type NetworkSummary } from '../api/networks'
import { Badge } from '../components/data-display/Badge'
import { DdnsStatusBadge } from '../components/data-display/DdnsStatusBadge'
import { EmptyState } from '../components/feedback/EmptyState'
import { ErrorState } from '../components/feedback/ErrorState'
import { LoadingState } from '../components/feedback/LoadingState'
import { formatDate } from '../utils/date'
import { errorMessage } from '../utils/errors'

export function NetworksPage() {
  const [networks, setNetworks] = useState<NetworkSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [message, setMessage] = useState('')
  const [deletingId, setDeletingId] = useState<number | null>(null)

  useEffect(() => {
    loadNetworks()
  }, [])

  async function loadNetworks() {
    setLoading(true)
    setMessage('')

    try {
      const nextNetworks = await listNetworks()
      setNetworks(nextNetworks)
    } catch (error) {
      setMessage(errorMessage(error, 'Unable to load networks.'))
    } finally {
      setLoading(false)
    }
  }

  async function handleDelete(network: NetworkSummary) {
    const confirmed = window.confirm(`Delete network "${network.name}"? This also removes its knock history.`)
    if (!confirmed) {
      return
    }

    setDeletingId(network.id)

    try {
      await deleteNetwork(network.id)
      await loadNetworks()
    } catch (error) {
      setMessage(errorMessage(error, 'Unable to delete the network.'))
      setLoading(false)
    } finally {
      setDeletingId(null)
    }
  }

  return (
    <section>
      <div className="page-header">
        <div>
          <h1>Networks</h1>
          <p>Track public IP changes and keep each DDNS target under one admin surface.</p>
        </div>
        <Link className="button" to="/admin/networks/new">
          New Network
        </Link>
      </div>

      {message && !loading && networks.length > 0 ? (
        <div className="panel-stack">
          <div className="form-message form-message-error">{message}</div>
        </div>
      ) : null}

      {loading ? <LoadingState label="Loading networks" /> : null}

      {!loading && message && networks.length === 0 ? (
        <ErrorState message={message} actionLabel="Retry" onAction={loadNetworks} />
      ) : null}

      {!loading && !message && networks.length === 0 ? (
        <section className="page-panel">
          <EmptyState
            title="No networks yet"
            message="Create the first managed network to start recording knocks and DDNS updates."
            action={
              <Link className="button" to="/admin/networks/new">
                Create Network
              </Link>
            }
          />
        </section>
      ) : null}

      {!loading && !message && networks.length > 0 ? (
        <>
          <section className="network-mobile-list">
            {networks.map((network) => (
              <article className="page-panel network-card" key={network.id}>
                <div className="network-card-header">
                  <div>
                    <Link className="table-link" to={`/admin/networks/${network.id}`}>
                      {network.name}
                    </Link>
                    <p>{formatDate(network.last_knock)}</p>
                  </div>
                  <DdnsStatusBadge status={network.ddns_status} />
                </div>

                <div className="network-card-grid">
                  <div>
                    <span className="meta-label">DDNS</span>
                    <Badge tone={network.ddns_enabled ? 'success' : 'neutral'}>
                      {network.ddns_enabled ? 'Enabled' : 'Disabled'}
                    </Badge>
                  </div>
                  <div>
                    <span className="meta-label">Provider</span>
                    <span>{network.ddns_type || 'None'}</span>
                  </div>
                  <div>
                    <span className="meta-label">Current IP</span>
                    <span className="mono">{network.current_ip ?? 'Unknown'}</span>
                  </div>
                  <div>
                    <span className="meta-label">Previous IP</span>
                    <span className="mono">{network.previous_ip ?? 'None'}</span>
                  </div>
                </div>

                <div className="inline-actions">
                  <Link className="button button-secondary" to={`/admin/networks/${network.id}`}>
                    Detail
                  </Link>
                  <Link className="button button-secondary" to={`/admin/networks/${network.id}/edit`}>
                    Edit
                  </Link>
                  <button
                    className="button button-secondary button-danger"
                    type="button"
                    disabled={deletingId === network.id}
                    onClick={() => handleDelete(network)}
                  >
                    {deletingId === network.id ? 'Deleting...' : 'Delete'}
                  </button>
                </div>
              </article>
            ))}
          </section>

          <section className="page-panel table-card desktop-network-table">
          <div className="table-wrap">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>DDNS</th>
                  <th>Provider</th>
                  <th>Current IP</th>
                  <th>Previous IP</th>
                  <th>Last Knock</th>
                  <th>Status</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {networks.map((network) => (
                  <tr key={network.id}>
                    <td>
                      <Link className="table-link" to={`/admin/networks/${network.id}`}>
                        {network.name}
                      </Link>
                    </td>
                    <td>
                      <Badge tone={network.ddns_enabled ? 'success' : 'neutral'}>
                        {network.ddns_enabled ? 'Enabled' : 'Disabled'}
                      </Badge>
                    </td>
                    <td>{network.ddns_type || 'None'}</td>
                    <td className="mono">{network.current_ip ?? 'Unknown'}</td>
                    <td className="mono">{network.previous_ip ?? 'None'}</td>
                    <td>{formatDate(network.last_knock)}</td>
                    <td>
                      <DdnsStatusBadge status={network.ddns_status} />
                    </td>
                    <td>
                      <div className="inline-actions">
                        <Link className="button button-secondary" to={`/admin/networks/${network.id}`}>
                          Detail
                        </Link>
                        <Link className="button button-secondary" to={`/admin/networks/${network.id}/edit`}>
                          Edit
                        </Link>
                        <button
                          className="button button-secondary button-danger"
                          type="button"
                          disabled={deletingId === network.id}
                          onClick={() => handleDelete(network)}
                        >
                          {deletingId === network.id ? 'Deleting...' : 'Delete'}
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          </section>
        </>
      ) : null}
    </section>
  )
}
