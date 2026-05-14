import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useI18n } from '../i18n'
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
  const { locale, t } = useI18n()
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

        setMessage(errorMessage(error, t('networkDetail.unableLoad')))
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

  async function handleDelete() {
    if (!networkId || !network) {
      return
    }

    const confirmed = window.confirm(t('networkDetail.deleteConfirm', { name: network.name }))
    if (!confirmed) {
      return
    }

    setDeleting(true)
    setMessage('')

    try {
      await deleteNetwork(networkId)
      navigate('/admin/networks', { replace: true })
    } catch (error) {
      setMessage(errorMessage(error, t('networkDetail.unableDelete')))
      setDeleting(false)
    }
  }

  async function handleCopy(value: string, kind: 'token' | 'curl' | 'crontab') {
    try {
      await navigator.clipboard.writeText(value)
      setCopyMessage(
        kind === 'token'
          ? t('networkDetail.copyTokenSuccess')
          : kind === 'curl'
            ? t('networkDetail.copyCurlSuccess')
            : t('networkDetail.copyCrontabSuccess'),
      )
    } catch {
      setCopyMessage(
        kind === 'token'
          ? t('networkDetail.copyTokenFailure')
          : kind === 'curl'
            ? t('networkDetail.copyCurlFailure')
            : t('networkDetail.copyCrontabFailure'),
      )
    }
  }

  async function handleRegenerateToken() {
    if (!networkId || !network) {
      return
    }

    const confirmed = window.confirm(
      t('networkDetail.regenerateConfirm', { name: network.name }),
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
      setCopyMessage(t('networkDetail.regenerateSuccess'))
    } catch (error) {
      setMessage(errorMessage(error, t('networkDetail.unableRegenerate')))
    } finally {
      setRegeneratingToken(false)
    }
  }

  if (!networkId) {
    return <ErrorState title={t('networkDetail.notFoundTitle')} message={t('networkDetail.notFoundMessage')} />
  }

  if (loading) {
    return <LoadingState label={t('networkDetail.loading')} />
  }

  if (notFound) {
    return <ErrorState title={t('networkDetail.notFoundTitle')} message={t('networkDetail.notFoundMessage')} />
  }

  if (message && !network) {
    return <ErrorState message={message} actionLabel={t('common.retry')} onAction={() => window.location.reload()} />
  }

  if (!network) {
    return <ErrorState message={t('networkDetail.unableLoad')} />
  }

  return (
    <section>
      <div className="page-header">
        <div>
          <h1>{network.name}</h1>
          <p>{t('networkDetail.subtitle')}</p>
        </div>
        <div className="page-actions">
          <Link className="button button-secondary" to={`/admin/networks/${network.id}/edit`}>
            {t('common.edit')}
          </Link>
          <Link className="button button-secondary" to={`/admin/networks/${network.id}/history`}>
            {t('common.history')}
          </Link>
          <button className="button button-secondary button-danger" type="button" disabled={deleting} onClick={handleDelete}>
            {deleting ? t('common.deleting') : t('common.delete')}
          </button>
        </div>
      </div>

      {message ? <div className="panel-stack"><div className="form-message form-message-error">{message}</div></div> : null}
      {copyMessage ? <div className="panel-stack"><div className="form-message form-message-success">{copyMessage}</div></div> : null}

      <section className="page-panel">
        <FieldList
          items={[
            { label: t('common.name'), value: network.name },
            {
              label: t('common.token'),
              value: (
                <div className="inline-actions">
                  <span className="mono">{network.token}</span>
                  <button className="button button-secondary" type="button" onClick={() => handleCopy(network.token, 'token')}>
                    {t('common.copy')}
                  </button>
                  <button
                    className="button button-secondary button-danger"
                    type="button"
                    disabled={regeneratingToken}
                    onClick={handleRegenerateToken}
                  >
                    {regeneratingToken ? t('common.regenerating') : t('common.regenerate')}
                  </button>
                </div>
              ),
            },
            {
              label: t('common.ddns'),
              value: (
                <Badge tone={network.ddns_enabled ? 'success' : 'neutral'}>
                  {network.ddns_enabled ? t('common.enabled') : t('common.disabled')}
                </Badge>
              ),
            },
            { label: t('common.provider'), value: network.ddns_type || t('common.none') },
            { label: t('common.currentIp'), value: network.current_ip ?? t('common.unknown'), mono: true },
            { label: t('common.previousIp'), value: network.previous_ip ?? t('common.none'), mono: true },
            { label: t('common.lastKnock'), value: formatDate(network.last_knock, locale, t('common.never')) },
            { label: t('networkDetail.latestDdnsStatus'), value: <DdnsStatusBadge status={network.ddns_status} /> },
          ]}
        />
      </section>

      <section className="page-panel command-panel">
        <div className="section-heading">
          <div>
            <h2>{t('networkDetail.clientCommandsTitle')}</h2>
            <p>{t('networkDetail.clientCommandsDescription')}</p>
          </div>
        </div>

        <div className="form-message form-message-info">
          <strong>{t('networkDetail.publicUrlTitle')}</strong>
          <p>{t('networkDetail.publicUrlMessage')}</p>
          <p className="mono">{network.commands.public_url}</p>
        </div>

        <div className="command-grid">
          <article className="command-card">
            <div className="command-card-header">
              <h3>curl</h3>
              <button className="button button-secondary" type="button" onClick={() => handleCopy(network.commands.curl, 'curl')}>
                {t('common.copy')}
              </button>
            </div>
            <pre className="code-block">
              <code>{network.commands.curl}</code>
            </pre>
          </article>

          <article className="command-card">
            <div className="command-card-header">
              <h3>crontab</h3>
              <button className="button button-secondary" type="button" onClick={() => handleCopy(network.commands.crontab, 'crontab')}>
                {t('common.copy')}
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
