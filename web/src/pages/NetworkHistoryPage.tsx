import { startTransition, useEffect, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
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

        setMessage(errorMessage(error, 'Unable to load knock history.'))
      })
      .finally(() => {
        if (active) {
          setLoading(false)
        }
      })

    return () => {
      active = false
    }
  }, [networkId, page, size])

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
    return <ErrorState title="Network not found" message="The requested network does not exist." />
  }

  if (loading) {
    return <LoadingState label="Loading knock history" />
  }

  if (notFound) {
    return <ErrorState title="Network not found" message="The requested network does not exist." />
  }

  if (message) {
    return <ErrorState message={message} actionLabel="Retry" onAction={() => window.location.reload()} />
  }

  return (
    <section>
      <div className="page-header">
        <div>
          <h1>Knock History</h1>
          <p>Inspect every request that updated or checked the public IP for {networkName}.</p>
        </div>
        <Link className="button button-secondary" to={`/admin/networks/${networkId}`}>
          Back to Detail
        </Link>
      </div>

      <section className="page-panel history-toolbar">
        <div className="history-summary">
          <strong>{total}</strong>
          <span>{total === 1 ? 'record' : 'records'}</span>
        </div>
        <label className="field field-inline">
          <span className="field-label">Page Size</span>
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
            Previous
          </button>
          <span>
            Page {page} / {totalPages}
          </span>
          <button
            className="button button-secondary"
            type="button"
            disabled={page >= totalPages}
            onClick={() => updatePagination(page + 1, size)}
          >
            Next
          </button>
        </div>
      </section>

      {records.length === 0 ? (
        <section className="page-panel">
          <EmptyState
            title="No history yet"
            message="Knock records will appear here after clients call the generated command."
            action={
              <Link className="button button-secondary" to={`/admin/networks/${networkId}`}>
                Back to Detail
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
                  <th>Time</th>
                  <th>IP</th>
                  <th>Previous IP</th>
                  <th>Changed</th>
                  <th>DDNS Status</th>
                  <th>DDNS Error</th>
                  <th>User-Agent</th>
                </tr>
              </thead>
              <tbody>
                {records.map((record) => (
                  <tr key={record.id}>
                    <td>{formatDate(record.created_at)}</td>
                    <td className="mono">{record.ip}</td>
                    <td className="mono">{record.previous_ip ?? 'None'}</td>
                    <td>
                      <Badge tone={record.ip_changed ? 'success' : 'neutral'}>{record.ip_changed ? 'Changed' : 'Same'}</Badge>
                    </td>
                    <td>
                      <DdnsStatusBadge status={record.ddns_status} />
                    </td>
                    <td className="history-error">{record.ddns_error || 'None'}</td>
                    <td className="history-agent">{record.user_agent || 'Unknown'}</td>
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
