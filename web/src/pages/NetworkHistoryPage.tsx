import { startTransition, useEffect, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { useI18n } from '../i18n'
import { getNetwork, listKnocks, type KnockRecord } from '../api/networks'
import { Badge } from '../components/data-display/Badge'
import { DdnsStatusBadge } from '../components/data-display/DdnsStatusBadge'
import { EmptyState } from '../components/feedback/EmptyState'
import { ErrorState } from '../components/feedback/ErrorState'
import { LoadingState } from '../components/feedback/LoadingState'
import { formatDate } from '../utils/date'
import { errorMessage, isApiError } from '../utils/errors'

const defaultPage = 1
const defaultSize = 50
const allowedPageSizes = new Set([20, 50, 100])

export function NetworkHistoryPage() {
  const { locale, t } = useI18n()
  const { networkId } = useParams()
  const [searchParams, setSearchParams] = useSearchParams()
  const [loading, setLoading] = useState(true)
  const [message, setMessage] = useState('')
  const [notFound, setNotFound] = useState(false)
  const [networkName, setNetworkName] = useState('')
  const [records, setRecords] = useState<KnockRecord[]>([])
  const [total, setTotal] = useState(0)

  const page = parsePage(searchParams.get('page'))
  const size = parseSize(searchParams.get('size'))
  const totalPages = Math.max(1, Math.ceil(total / size))

  useEffect(() => {
    if (!networkId) {
      return
    }

    let active = true

    Promise.all([getNetwork(networkId), listKnocks(networkId, page, size)])
      .then(([network, history]) => {
        if (!active) {
          return
        }

        setNetworkName(network.name)
        setRecords(history.records)
        setTotal(history.total)
      })
      .catch((error) => {
        if (!active) {
          return
        }

        if (isApiError(error) && error.status === 404) {
          setNotFound(true)
          return
        }

        setMessage(errorMessage(error, t('networkHistory.unableLoad')))
      })
      .finally(() => {
        if (active) {
          setLoading(false)
        }
      })

    return () => {
      active = false
    }
  }, [networkId, page, size, t])

  function updatePagination(nextPage: number, nextSize: number) {
    const params = new URLSearchParams(searchParams)
    params.set('page', String(nextPage))
    params.set('size', String(nextSize))

    setLoading(true)
    setMessage('')
    setNotFound(false)

    startTransition(() => {
      setSearchParams(params)
    })
  }

  if (!networkId) {
    return <ErrorState title={t('networkHistory.notFoundTitle')} message={t('networkHistory.notFoundMessage')} />
  }

  if (loading) {
    return <LoadingState label={t('networkHistory.loading')} />
  }

  if (notFound) {
    return <ErrorState title={t('networkHistory.notFoundTitle')} message={t('networkHistory.notFoundMessage')} />
  }

  if (message) {
    return <ErrorState message={message} actionLabel={t('common.retry')} onAction={() => window.location.reload()} />
  }

  return (
    <section>
      <div className="page-header">
        <div>
          <h1>{t('networkHistory.title')}</h1>
          <p>{t('networkHistory.subtitle', { name: networkName })}</p>
        </div>
        <Link className="button button-secondary" to={`/admin/networks/${networkId}`}>
          {t('common.backToDetail')}
        </Link>
      </div>

      <section className="page-panel history-toolbar">
        <div className="history-summary">
          <strong>{total}</strong>
          <span>{total === 1 ? t('networkHistory.recordSingular') : t('networkHistory.recordPlural')}</span>
        </div>
        <label className="field field-inline">
          <span className="field-label">{t('common.pageSize')}</span>
          <select
            className="field-select"
            value={size}
            onChange={(event) => updatePagination(defaultPage, parseSize(event.target.value))}
          >
            {[20, 50, 100].map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </label>
        <div className="pagination">
          <button className="button button-secondary" type="button" disabled={page <= 1} onClick={() => updatePagination(page - 1, size)}>
            {t('common.previous')}
          </button>
          <span>{t('networkHistory.pageSummary', { page, totalPages })}</span>
          <button
            className="button button-secondary"
            type="button"
            disabled={page >= totalPages}
            onClick={() => updatePagination(page + 1, size)}
          >
            {t('common.next')}
          </button>
        </div>
      </section>

      {records.length === 0 ? (
        <section className="page-panel">
          <EmptyState
            title={t('networkHistory.noHistoryTitle')}
            message={t('networkHistory.noHistoryMessage')}
            action={
              <Link className="button button-secondary" to={`/admin/networks/${networkId}`}>
                {t('common.backToDetail')}
              </Link>
            }
          />
        </section>
      ) : (
        <section className="page-panel table-card">
          <div className="table-wrap">
            <table className="data-table history-table">
              <thead>
                <tr>
                  <th>{t('common.time')}</th>
                  <th>IP</th>
                  <th>{t('common.previousIp')}</th>
                  <th>{t('common.changed')}</th>
                  <th>{t('common.ddnsStatus')}</th>
                  <th>{t('common.ddnsError')}</th>
                  <th>{t('common.userAgent')}</th>
                </tr>
              </thead>
              <tbody>
                {records.map((record) => (
                  <tr key={record.id}>
                    <td>{formatDate(record.created_at, locale, t('common.never'))}</td>
                    <td className="mono">{record.ip}</td>
                    <td className="mono">{record.previous_ip ?? t('common.none')}</td>
                    <td>
                      <Badge tone={record.ip_changed ? 'success' : 'neutral'}>
                        {record.ip_changed ? t('common.changed') : t('common.same')}
                      </Badge>
                    </td>
                    <td>
                      <DdnsStatusBadge status={record.ddns_status} />
                    </td>
                    <td className="history-error">{record.ddns_error || t('common.none')}</td>
                    <td className="history-agent">{record.user_agent || t('common.unknown')}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}
    </section>
  )
}

function parsePage(value: string | null): number {
  const parsed = Number(value ?? defaultPage)
  return Number.isInteger(parsed) && parsed > 0 ? parsed : defaultPage
}

function parseSize(value: string | null): number {
  const parsed = Number(value ?? defaultSize)
  return Number.isInteger(parsed) && allowedPageSizes.has(parsed) ? parsed : defaultSize
}
