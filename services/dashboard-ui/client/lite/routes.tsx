import type { NonIndexRouteObject, RouteObject } from 'react-router'
import { PageTransition } from './components/templates/PageTransition'
import { FocusLayout } from './pages/FocusLayout'
import { AppBranchActivity } from './pages/AppBranchActivity'
import { AppBranchConfig } from './pages/AppBranchConfig'
import { AppBranchLayout } from './pages/AppBranchLayout'
import { AppBranchOverview } from './pages/AppBranchOverview'
import { AppLayout } from './pages/AppLayout'
import { AppResolver } from './pages/AppResolver'
import { AppSetup } from './pages/AppSetup'
import { Apps } from './pages/Apps'
import { InstallActivity } from './pages/InstallActivity'
import { InstallLayout } from './pages/InstallLayout'
import { InstallOverview } from './pages/InstallOverview'
import { Installs } from './pages/Installs'
import { InstallSetup } from './pages/InstallSetup'
import { ModuleLayout } from './pages/ModuleLayout'
import { Modules } from './pages/Modules'
import { Team } from './pages/Team'
import { OrgLayout } from './pages/OrgLayout'
import { RootLayout } from './pages/RootLayout'
import { SettingsLayout } from './pages/SettingsLayout'
import {
  ApiTokens,
  Connections,
  Dashboard,
  NotFound,
  OidcFederation,
  Onboarding,
  ServiceAccounts,
  Triggers,
  Webhooks,
} from './pages/scaffolds'
import { SETTINGS_MODULE_IDS } from './utils/modules'

const isLayoutRoute = (
  route: RouteObject
): route is NonIndexRouteObject & { children: RouteObject[] } =>
  !!route.children

export const withPageTransitions = (routes: RouteObject[]): RouteObject[] =>
  routes.map((route): RouteObject => {
    if (isLayoutRoute(route)) {
      return { ...route, children: withPageTransitions(route.children) }
    }
    if (!route.element) return route

    return {
      ...route,
      element: <PageTransition>{route.element}</PageTransition>,
    }
  })

export const liteRoutes: RouteObject[] = withPageTransitions([
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
          {
            id: 'module-apps',
            element: <ModuleLayout module="apps" />,
            children: [
              { id: 'apps', path: 'apps', element: <Apps /> },
              {
                id: 'app-setup',
                path: 'apps/setup',
                element: <AppSetup />,
              },
              {
                id: 'app-layout',
                path: 'apps/:appId',
                element: <AppLayout />,
                children: [
                  {
                    id: 'app-resolver',
                    index: true,
                    element: <AppResolver />,
                  },
                  {
                    id: 'app-branch-layout',
                    path: 'branches/:branchId',
                    element: <AppBranchLayout />,
                    children: [
                      {
                        id: 'app-branch-overview',
                        index: true,
                        element: <AppBranchOverview />,
                      },
                      {
                        id: 'app-branch-activity',
                        path: 'activity',
                        element: <AppBranchActivity />,
                      },
                      {
                        id: 'app-branch-config',
                        path: 'config',
                        element: <AppBranchConfig />,
                      },
                    ],
                  },
                ],
              },
            ],
          },
          {
            id: 'module-installs',
            element: <ModuleLayout module="installs" />,
            children: [
              { id: 'installs', path: 'installs', element: <Installs /> },
              {
                id: 'install-setup',
                path: 'installs/setup',
                element: <InstallSetup />,
              },
              {
                id: 'install-layout',
                path: 'installs/:installId',
                element: <InstallLayout />,
                children: [
                  {
                    id: 'install-overview',
                    index: true,
                    element: <InstallOverview />,
                  },
                  {
                    id: 'install-activity',
                    path: 'activity',
                    element: <InstallActivity />,
                  },
                ],
              },
            ],
          },
          {
            id: 'module-team',
            element: <ModuleLayout module="team" />,
            children: [{ id: 'teams', path: 'teams', element: <Team /> }],
          },
          { id: 'modules', path: 'modules', element: <Modules /> },
          {
            id: 'module-settings',
            element: <ModuleLayout module={SETTINGS_MODULE_IDS} />,
            children: [
              {
                id: 'settings-layout',
                path: 'settings',
                element: <SettingsLayout />,
                children: [
                  {
                    id: 'module-connections',
                    element: <ModuleLayout module="connections" />,
                    children: [
                      {
                        id: 'settings-connections',
                        index: true,
                        element: <Connections />,
                      },
                    ],
                  },
                  {
                    id: 'module-webhooks',
                    element: <ModuleLayout module="webhooks" />,
                    children: [
                      {
                        id: 'settings-webhooks',
                        path: 'webhooks',
                        element: <Webhooks />,
                      },
                    ],
                  },
                  {
                    id: 'module-triggers',
                    element: <ModuleLayout module="triggers" />,
                    children: [
                      {
                        id: 'settings-triggers',
                        path: 'triggers',
                        element: <Triggers />,
                      },
                    ],
                  },
                  {
                    id: 'module-api-tokens',
                    element: <ModuleLayout module="api-tokens" />,
                    children: [
                      {
                        id: 'settings-api-tokens',
                        path: 'api-tokens',
                        element: <ApiTokens />,
                      },
                    ],
                  },
                  {
                    id: 'module-service-accounts',
                    element: <ModuleLayout module="service-accounts" />,
                    children: [
                      {
                        id: 'settings-service-accounts',
                        path: 'service-accounts',
                        element: <ServiceAccounts />,
                      },
                    ],
                  },
                  {
                    id: 'module-oidc-federation',
                    element: <ModuleLayout module="oidc-federation" />,
                    children: [
                      {
                        id: 'settings-oidc',
                        path: 'oidc',
                        element: <OidcFederation />,
                      },
                    ],
                  },
                ],
              },
            ],
          },
          { id: 'org-not-found', path: '*', element: <NotFound /> },
        ],
      },
    ],
  },
])
