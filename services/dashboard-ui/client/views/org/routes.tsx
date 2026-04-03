import type { RouteObject } from 'react-router'
import { OrgLayout } from './OrgLayout'
import { Dashboard } from './Dashbaord'
import { Apps } from './Apps'
import { Installs } from './Installs'
import { BuildRunner } from './BuildRunner'
import { RunnerJobDetail } from './RunnerJobDetail'
import { RunnerProcesses } from './RunnerProcesses'
import { ProcessSystemLogs } from './ProcessSystemLogs'
import { Team } from './Team'
import { NotFound } from '@/views/NotFound'
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
      { path: ':orgId/runner/jobs/:jobId', element: <RunnerJobDetail /> },
      { path: ':orgId/runner/processes', element: <RunnerProcesses /> },
      { path: ':orgId/runner/processes/:processId/logs', element: <ProcessSystemLogs /> },
      { path: ':orgId/team', element: <Team /> },
      ...appRoutes,
      ...installRoutes,
      { path: ':orgId/*', element: <NotFound /> },
    ],
  },
]
