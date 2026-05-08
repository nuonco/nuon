import { redirect, type RouteObject } from 'react-router'
import { OrgLayout } from './OrgLayout'
import { Dashboard } from './Dashbaord'
import { Apps } from './Apps'
import { Installs } from './Installs'
import { BuildRunner } from './BuildRunner'
import { RunnerJobDetail } from './RunnerJobDetail'
import { RunnerProcesses } from './RunnerProcesses'
import { ProcessSystemLogs } from './ProcessSystemLogs'
import { Team } from './Team'
import { VCSConnectionDetail } from './VCSConnectionDetail'
import { Slack } from './Slack'
import { Webhooks } from './Webhooks'
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
      { path: ':orgId/webhooks', element: <Webhooks /> },
      { path: ':orgId/slack', element: <Slack /> },
      { path: ':orgId/connections', loader: ({ params }) => redirect(`/${params.orgId}`) },
      { path: ':orgId/connections/vcs', loader: ({ params }) => redirect(`/${params.orgId}`) },
      { path: ':orgId/connections/vcs/:connectionId', element: <VCSConnectionDetail /> },
      { path: ':orgId/connections/:connectionId', loader: ({ params }) => redirect(`/${params.orgId}/connections/vcs/${params.connectionId}`) },
      ...appRoutes,
      ...installRoutes,
      { path: ':orgId/*', element: <NotFound /> },
    ],
  },
]
