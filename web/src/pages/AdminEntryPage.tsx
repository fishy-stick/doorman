import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { checkSession } from '../api/auth'
import { ErrorState } from '../components/feedback/ErrorState'
import { LoadingState } from '../components/feedback/LoadingState'
import { errorMessage, isUnauthorized } from '../utils/errors'

export function AdminEntryPage() {
  const navigate = useNavigate()
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

        setMessage(errorMessage(error, 'Unable to check the admin session.'))
      })

    return () => {
      active = false
    }
  }, [navigate])

  if (message) {
    return (
      <main className="screen">
        <ErrorState message={message} actionLabel="Retry" onAction={() => window.location.reload()} />
      </main>
    )
  }

  return <LoadingState label="Opening admin" />
}
