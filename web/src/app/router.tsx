import { useEffect, useState } from 'react'
import { createBrowserRouter, Navigate, Outlet, useNavigate } from 'react-router-dom'
import { checkSession } from '../api/auth'
import { ProtectedLayout } from '../components/layout/AdminLayout'
import { LoadingState } from '../components/feedback/LoadingState'
import { ErrorState } from '../components/feedback/ErrorState'
import { ChangePasswordPage } from '../pages/ChangePasswordPage'
import { LoginPage } from '../pages/LoginPage'
import { NetworkDetailPage } from '../pages/NetworkDetailPage'
import { NetworkEditPage } from '../pages/NetworkEditPage'
import { NetworkHistoryPage } from '../pages/NetworkHistoryPage'
import { NetworkNewPage } from '../pages/NetworkNewPage'
import { NetworksPage } from '../pages/NetworksPage'
import { NotFoundPage } from '../pages/NotFoundPage'
import { errorMessage, isUnauthorized } from '../utils/errors'

export const router = createBrowserRouter([
  {
    path: '/admin',
    element: <Outlet />,
    children: [
      {
        index: true,
        element: <RootRedirect />,
      },
      {
        path: 'login',
        element: <LoginPage />,
      },
      {
        element: <ProtectedLayout />,
        children: [
          {
            path: 'networks',
            element: <NetworksPage />,
          },
          {
            path: 'networks/new',
            element: <NetworkNewPage />,
          },
          {
            path: 'networks/:networkId',
            element: <NetworkDetailPage />,
          },
          {
            path: 'networks/:networkId/edit',
            element: <NetworkEditPage />,
          },
          {
            path: 'networks/:networkId/history',
            element: <NetworkHistoryPage />,
          },
          {
            path: 'settings/password',
            element: <ChangePasswordPage />,
          },
          {
            path: '*',
            element: <NotFoundPage />,
          },
        ],
      },
    ],
  },
  {
    path: '*',
    element: <Navigate to="/admin" replace />,
  },
])

function RootRedirect() {
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
