import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { checkSession, login } from '../api/auth'
import { ErrorState } from '../components/feedback/ErrorState'
import { LoadingState } from '../components/feedback/LoadingState'
import { errorMessage, isUnauthorized } from '../utils/errors'

export function LoginPage() {
  const navigate = useNavigate()
  const [password, setPassword] = useState('')
  const [checkingSession, setCheckingSession] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [message, setMessage] = useState('')
  const [fatalError, setFatalError] = useState('')

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

        if (!isUnauthorized(error)) {
          setFatalError(errorMessage(error, 'Unable to check the admin session.'))
        }
      })
      .finally(() => {
        if (active) {
          setCheckingSession(false)
        }
      })

    return () => {
      active = false
    }
  }, [navigate])

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const trimmedPassword = password.trim()
    if (trimmedPassword === '') {
      setMessage('Password is required.')
      return
    }

    setSubmitting(true)
    setMessage('')

    try {
      await login(trimmedPassword)
      navigate('/admin/networks', { replace: true })
    } catch (error) {
      setMessage(errorMessage(error, 'Login failed.'))
    } finally {
      setSubmitting(false)
    }
  }

  if (checkingSession) {
    return (
      <main className="screen">
        <LoadingState label="Checking session" />
      </main>
    )
  }

  if (fatalError) {
    return (
      <main className="screen">
        <ErrorState message={fatalError} actionLabel="Retry" onAction={() => window.location.reload()} />
      </main>
    )
  }

  return (
    <main className="screen login-screen">
      <section className="page-panel auth-panel">
        <div className="page-header auth-header">
          <div>
            <h1>Doorman</h1>
            <p>Sign in to manage networks, DDNS settings, and knock history.</p>
          </div>
        </div>
        <form className="form-stack" onSubmit={handleSubmit}>
          <label className="field">
            <span className="field-label">Password</span>
            <input
              autoComplete="current-password"
              className="field-input"
              name="password"
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              disabled={submitting}
            />
          </label>
          {message ? <div className="form-message form-message-error">{message}</div> : null}
          <button className="button button-block" type="submit" disabled={submitting}>
            {submitting ? 'Signing in...' : 'Sign In'}
          </button>
        </form>
      </section>
    </main>
  )
}
