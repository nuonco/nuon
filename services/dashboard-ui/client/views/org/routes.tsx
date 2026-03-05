import type { RouteObject } from 'react-router'
import { OrgLayout } from './OrgLayout'
import { Dashboard } from './Dashbaord'
import { Apps } from './Apps'
import { Installs } from './Installs'
import { BuildRunner } from './BuildRunner'
import { Team } from './Team'
import { appRoutes } from '@/views/app/routes'
import { installRoutes } from '@/views/install/routes'

export const orgRoutes: RouteObject[] = [
  {
    element: <OrgLayout />,
    children: [
      { path: ':orgId', element: <Dashboard /> },
      { path: ':orgId/apps', element: <Apps /> },
      { path: ':orgId/installs', element: <Installs /> },
      { path: ':orgId/runner', element: <BuildRunner /> },
      { path: ':orgId/team', element: <Team /> },
      ...appRoutes,
      ...installRoutes,
    ],
  },
]
