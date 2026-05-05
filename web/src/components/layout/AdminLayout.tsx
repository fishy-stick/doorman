import { useEffect, useState } from 'react'
import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { checkSession, logout } from '../../api/auth'
import { setUnauthorizedHandler } from '../../api/client'
import { useI18n } from '../../i18n'
import { ErrorState } from '../feedback/ErrorState'
import { LoadingState } from '../feedback/LoadingState'
import { errorMessage, isUnauthorized } from '../../utils/errors'
import { LanguageSwitcher } from './LanguageSwitcher'

export function ProtectedLayout() {
  const navigate = useNavigate()
  const { t } = useI18n()
  const [status, setStatus] = useState<'checking' | 'ready' | 'error'>('checking')
  const [message, setMessage] = useState('')

  useEffect(() => {
    let active = true

    setUnauthorizedHandler(() => {
      navigate('/admin/login', { replace: true })
    })

    checkSession({ skipUnauthorized: true })
      .then(() => {
        if (active) {
          setStatus('ready')
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
        setStatus('error')
      })

    return () => {
      active = false
      setUnauthorizedHandler(null)
    }
  }, [navigate, t])

  if (status === 'checking') {
    return <LoadingState label={t('login.checkingSession')} />
  }

  if (status === 'error') {
    return (
      <main className="screen">
        <ErrorState
          message={message}
          actionLabel={t('common.retry')}
          onAction={() => {
            setStatus('checking')
            window.location.reload()
          }}
        />
      </main>
    )
  }

  return <AdminShell />
}

function AdminShell() {
  const navigate = useNavigate()
  const { t } = useI18n()
  const [logoutError, setLogoutError] = useState('')
  const [loggingOut, setLoggingOut] = useState(false)

  async function handleLogout() {
    setLoggingOut(true)
    setLogoutError('')

    try {
      await logout()
      navigate('/admin/login', { replace: true })
    } catch (error) {
      setLogoutError(errorMessage(error, t('layout.logoutFailed')))
      navigate('/admin/login', { replace: true })
    } finally {
      setLoggingOut(false)
    }
  }

  return (
    <div className="admin-shell">
      <header className="topbar">
        <NavLink className="brand" to="/admin/networks">
          <span className="brand-mark">D</span>
          <span>
            <strong>Doorman</strong>
            <small>{t('layout.networkAdmin')}</small>
          </span>
        </NavLink>
        <nav className="nav-links" aria-label={t('layout.adminNavigation')}>
          <NavLink to="/admin/networks">{t('common.networks')}</NavLink>
          <NavLink to="/admin/settings/password">{t('layout.changePassword')}</NavLink>
          <LanguageSwitcher />
          <button className="nav-button" type="button" onClick={handleLogout} disabled={loggingOut}>
            {loggingOut ? t('layout.loggingOut') : t('layout.logout')}
          </button>
        </nav>
      </header>
      {logoutError ? <div className="toast-error">{logoutError}</div> : null}
      <main className="content">
        <Outlet />
      </main>
    </div>
  )
}
