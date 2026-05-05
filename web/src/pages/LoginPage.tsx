import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useI18n } from '../i18n'
import { checkSession, login } from '../api/auth'
import { LanguageSwitcher } from '../components/layout/LanguageSwitcher'
import { ErrorState } from '../components/feedback/ErrorState'
import { LoadingState } from '../components/feedback/LoadingState'
import { errorMessage, isUnauthorized } from '../utils/errors'

export function LoginPage() {
  const navigate = useNavigate()
  const { t } = useI18n()
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
          setFatalError(errorMessage(error, t('login.unableCheckSession')))
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
  }, [navigate, t])

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const trimmedPassword = password.trim()
    if (trimmedPassword === '') {
      setMessage(t('login.passwordRequired'))
      return
    }

    setSubmitting(true)
    setMessage('')

    try {
      await login(trimmedPassword)
      navigate('/admin/networks', { replace: true })
    } catch (error) {
      setMessage(errorMessage(error, t('login.loginFailed')))
    } finally {
      setSubmitting(false)
    }
  }

  if (checkingSession) {
    return (
      <main className="screen">
        <LoadingState label={t('login.checkingSession')} />
      </main>
    )
  }

  if (fatalError) {
    return (
      <main className="screen">
        <ErrorState message={fatalError} actionLabel={t('common.retry')} onAction={() => window.location.reload()} />
      </main>
    )
  }

  return (
    <main className="screen login-screen">
      <section className="page-panel auth-panel">
        <div className="auth-toolbar">
          <LanguageSwitcher />
        </div>
        <div className="page-header auth-header">
          <div>
            <h1>Doorman</h1>
            <p>{t('login.description')}</p>
          </div>
        </div>
        <form className="form-stack" onSubmit={handleSubmit}>
          <label className="field">
            <span className="field-label">{t('login.password')}</span>
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
            {submitting ? t('login.signingIn') : t('login.signIn')}
          </button>
        </form>
      </section>
    </main>
  )
}
