import { createBrowserRouter, Navigate, Outlet } from 'react-router-dom'
import { ProtectedLayout } from '../components/layout/AdminLayout'
import { AdminEntryPage } from '../pages/AdminEntryPage'
import { ChangePasswordPage } from '../pages/ChangePasswordPage'
import { LoginPage } from '../pages/LoginPage'
import { NetworkDetailPage } from '../pages/NetworkDetailPage'
import { NetworkEditPage } from '../pages/NetworkEditPage'
import { NetworkHistoryPage } from '../pages/NetworkHistoryPage'
import { NetworkNewPage } from '../pages/NetworkNewPage'
import { NetworksPage } from '../pages/NetworksPage'
import { NotFoundPage } from '../pages/NotFoundPage'

export const router = createBrowserRouter([
  {
    path: '/admin',
    element: <Outlet />,
    children: [
      {
        index: true,
        element: <AdminEntryPage />,
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
