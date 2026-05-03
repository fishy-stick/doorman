import { useEffect, useState } from 'react'
import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { checkSession, logout } from '../../api/auth'
import { setUnauthorizedHandler } from '../../api/client'
import { ErrorState } from '../feedback/ErrorState'
import { LoadingState } from '../feedback/LoadingState'
import { errorMessage, isUnauthorized } from '../../utils/errors'

export function ProtectedLayout() {
  const navigate = useNavigate()
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

        setMessage(errorMessage(error, 'Unable to verify the admin session.'))
        setStatus('error')
      })

    return () => {
      active = false
      setUnauthorizedHandler(null)
    }
  }, [navigate])

  if (status === 'checking') {
    return <LoadingState label="Checking session" />
  }

  if (status === 'error') {
    return (
      <main className="screen">
        <ErrorState
          message={message}
          actionLabel="Retry"
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
  const [logoutError, setLogoutError] = useState('')
  const [loggingOut, setLoggingOut] = useState(false)

  async function handleLogout() {
    setLoggingOut(true)
    setLogoutError('')

    try {
      await logout()
      navigate('/admin/login', { replace: true })
    } catch (error) {
      setLogoutError(errorMessage(error, 'Logout failed.'))
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
            <small>Network admin</small>
          </span>
        </NavLink>
        <nav className="nav-links" aria-label="Admin navigation">
          <NavLink to="/admin/networks">Networks</NavLink>
          <NavLink to="/admin/settings/password">Change Password</NavLink>
          <button className="nav-button" type="button" onClick={handleLogout} disabled={loggingOut}>
            {loggingOut ? 'Logging out' : 'Logout'}
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
