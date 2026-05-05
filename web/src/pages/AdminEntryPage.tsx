import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useI18n } from '../i18n'
import { checkSession } from '../api/auth'
import { ErrorState } from '../components/feedback/ErrorState'
import { LoadingState } from '../components/feedback/LoadingState'
import { errorMessage, isUnauthorized } from '../utils/errors'

export function AdminEntryPage() {
  const navigate = useNavigate()
  const { t } = useI18n()
  const [message, setMessage] = useState('')

  useEffect(() => {
    let active = true

    checkSession({ skipUnauthorized: true })
      .then(() => {
        if (active) {
          navigate('/admin/networks', { replace: true })
        }
      })
      .catch((error: unknown) => {
        if (!active) {
          return
        }

        if (isUnauthorized(error)) {
          navigate('/admin/login', { replace: true })
          return
        }

        setMessage(errorMessage(error, t('adminEntry.unableCheckSession')))
      })

    return () => {
      active = false
    }
  }, [navigate, t])

  if (message) {
    return (
      <main className="screen">
        <ErrorState message={message} actionLabel={t('common.retry')} onAction={() => window.location.reload()} />
      </main>
    )
  }

  return <LoadingState label={t('adminEntry.openingAdmin')} />
}
