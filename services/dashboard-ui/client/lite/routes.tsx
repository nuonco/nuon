import type { RouteObject } from 'react-router'
import { FocusLayout } from './pages/FocusLayout'
import { OrgLayout } from './pages/OrgLayout'
import { RootLayout } from './pages/RootLayout'
import { SettingsLayout } from './pages/SettingsLayout'
import {
  ApiTokens,
  Apps,
  Connections,
  Dashboard,
  Installs,
  NotFound,
  OidcFederation,
  Onboarding,
  ServiceAccounts,
  Teams,
  Triggers,
  Webhooks,
} from './pages/scaffolds'

export const liteRoutes: RouteObject[] = [
  {
    id: 'root-layout',
    element: <RootLayout />,
    children: [
      {
        id: 'focus-layout',
        element: <FocusLayout />,
        children: [
          {
            id: 'onboarding',
            path: 'onboarding',
            element: <Onboarding />,
          },
        ],
      },
      {
        id: 'org-layout',
        path: ':orgId',
        element: <OrgLayout />,
        children: [
          { id: 'dashboard', index: true, element: <Dashboard /> },
          { id: 'apps', path: 'apps', element: <Apps /> },
          { id: 'installs', path: 'installs', element: <Installs /> },
          { id: 'teams', path: 'teams', element: <Teams /> },
          {
            id: 'settings-layout',
            path: 'settings',
            element: <SettingsLayout />,
            children: [
              {
                id: 'settings-connections',
                index: true,
                element: <Connections />,
              },
              {
                id: 'settings-webhooks',
                path: 'webhooks',
                element: <Webhooks />,
              },
              {
                id: 'settings-triggers',
                path: 'triggers',
                element: <Triggers />,
              },
              {
                id: 'settings-api-tokens',
                path: 'api-tokens',
                element: <ApiTokens />,
              },
              {
                id: 'settings-service-accounts',
                path: 'service-accounts',
                element: <ServiceAccounts />,
              },
              {
                id: 'settings-oidc',
                path: 'oidc',
                element: <OidcFederation />,
              },
            ],
          },
          { id: 'org-not-found', path: '*', element: <NotFound /> },
        ],
      },
    ],
  },
]
