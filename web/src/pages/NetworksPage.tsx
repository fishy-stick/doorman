import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useI18n } from '../i18n'
import { deleteNetwork, listNetworks, type NetworkSummary } from '../api/networks'
import { Badge } from '../components/data-display/Badge'
import { DdnsStatusBadge } from '../components/data-display/DdnsStatusBadge'
import { EmptyState } from '../components/feedback/EmptyState'
import { ErrorState } from '../components/feedback/ErrorState'
import { LoadingState } from '../components/feedback/LoadingState'
import { formatDate } from '../utils/date'
import { errorMessage } from '../utils/errors'

export function NetworksPage() {
  const { locale, t } = useI18n()
  const [networks, setNetworks] = useState<NetworkSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [message, setMessage] = useState('')
  const [deletingId, setDeletingId] = useState<number | null>(null)

  async function loadNetworks() {
    setLoading(true)
    setMessage('')

    try {
      const nextNetworks = await listNetworks()
      setNetworks(nextNetworks)
    } catch (error) {
      setMessage(errorMessage(error, t('networks.unableLoad')))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    let active = true

    async function loadInitialNetworks() {
      try {
        const nextNetworks = await listNetworks()
        if (active) {
          setNetworks(nextNetworks)
        }
      } catch (error) {
        if (active) {
          setMessage(errorMessage(error, t('networks.unableLoad')))
        }
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }

    void loadInitialNetworks()

    return () => {
      active = false
    }
  }, [t])

  async function handleDelete(network: NetworkSummary) {
    const confirmed = window.confirm(t('networks.deleteConfirm', { name: network.name }))
    if (!confirmed) {
      return
    }

    setDeletingId(network.id)

    try {
      await deleteNetwork(network.id)
      await loadNetworks()
    } catch (error) {
      setMessage(errorMessage(error, t('networks.unableDelete')))
      setLoading(false)
    } finally {
      setDeletingId(null)
    }
  }

  return (
    <section>
      <div className="page-header">
        <div>
          <h1>{t('networks.title')}</h1>
          <p>{t('networks.subtitle')}</p>
        </div>
        <Link className="button" to="/admin/networks/new">
          {t('networks.newNetwork')}
        </Link>
      </div>

      {message && !loading && networks.length > 0 ? (
        <div className="panel-stack">
          <div className="form-message form-message-error">{message}</div>
        </div>
      ) : null}

      {loading ? <LoadingState label={t('networks.loading')} /> : null}

      {!loading && message && networks.length === 0 ? (
        <ErrorState message={message} actionLabel={t('common.retry')} onAction={loadNetworks} />
      ) : null}

      {!loading && !message && networks.length === 0 ? (
        <section className="page-panel">
          <EmptyState
            title={t('networks.noNetworksTitle')}
            message={t('networks.noNetworksMessage')}
            action={
              <Link className="button" to="/admin/networks/new">
                {t('networks.createNetwork')}
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
                    <p>{formatDate(network.last_knock, locale, t('common.never'))}</p>
                  </div>
                  <DdnsStatusBadge status={network.ddns_status} />
                </div>

                <div className="network-card-grid">
                  <div>
                    <span className="meta-label">{t('common.ddns')}</span>
                    <Badge tone={network.ddns_enabled ? 'success' : 'neutral'}>
                      {network.ddns_enabled ? t('common.enabled') : t('common.disabled')}
                    </Badge>
                  </div>
                  <div>
                    <span className="meta-label">{t('common.provider')}</span>
                    <span>{network.ddns_type || t('common.none')}</span>
                  </div>
                  <div>
                    <span className="meta-label">{t('common.currentIp')}</span>
                    <span className="mono">{network.current_ip ?? t('common.unknown')}</span>
                  </div>
                  <div>
                    <span className="meta-label">{t('common.previousIp')}</span>
                    <span className="mono">{network.previous_ip ?? t('common.none')}</span>
                  </div>
                </div>

                <div className="inline-actions">
                  <Link className="button button-secondary" to={`/admin/networks/${network.id}`}>
                    {t('common.detail')}
                  </Link>
                  <Link className="button button-secondary" to={`/admin/networks/${network.id}/edit`}>
                    {t('common.edit')}
                  </Link>
                  <button
                    className="button button-secondary button-danger"
                    type="button"
                    disabled={deletingId === network.id}
                    onClick={() => handleDelete(network)}
                  >
                    {deletingId === network.id ? t('common.deleting') : t('common.delete')}
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
                  <th>{t('common.name')}</th>
                  <th>{t('common.ddns')}</th>
                  <th>{t('common.provider')}</th>
                  <th>{t('common.currentIp')}</th>
                  <th>{t('common.previousIp')}</th>
                  <th>{t('common.lastKnock')}</th>
                  <th>{t('common.status')}</th>
                  <th>{t('common.actions')}</th>
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
                        {network.ddns_enabled ? t('common.enabled') : t('common.disabled')}
                      </Badge>
                    </td>
                    <td>{network.ddns_type || t('common.none')}</td>
                    <td className="mono">{network.current_ip ?? t('common.unknown')}</td>
                    <td className="mono">{network.previous_ip ?? t('common.none')}</td>
                    <td>{formatDate(network.last_knock, locale, t('common.never'))}</td>
                    <td>
                      <DdnsStatusBadge status={network.ddns_status} />
                    </td>
                    <td>
                      <div className="inline-actions">
                        <Link className="button button-secondary" to={`/admin/networks/${network.id}`}>
                          {t('common.detail')}
                        </Link>
                        <Link className="button button-secondary" to={`/admin/networks/${network.id}/edit`}>
                          {t('common.edit')}
                        </Link>
                        <button
                          className="button button-secondary button-danger"
                          type="button"
                          disabled={deletingId === network.id}
                          onClick={() => handleDelete(network)}
                        >
                          {deletingId === network.id ? t('common.deleting') : t('common.delete')}
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
