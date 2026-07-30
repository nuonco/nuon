import { Outlet, redirect, type RouteObject } from 'react-router'
import { OrgLayout } from './OrgLayout'
import { Dashboard } from './Dashbaord'
import { Apps } from './Apps'
import { Installs } from './Installs'
import { BuildRunner } from './BuildRunner'
import { RunnerJobDetail } from './RunnerJobDetail'
import { RunnerProcesses } from './RunnerProcesses'
import { ProcessSystemLogs } from './ProcessSystemLogs'
import { Team } from './Team'
import { ApiTokens } from './ApiTokens'
import { ServiceAccounts } from './ServiceAccounts'
import { VCSConnectionDetail } from './VCSConnectionDetail'
import { Slack } from './Slack'
import { Webhooks } from './Webhooks'
import { Triggers } from './Triggers'
import { TriggerLayout } from './TriggerLayout'
import { TriggerOverview } from './trigger-tabs/TriggerOverview'
import { TriggerRules } from './trigger-tabs/TriggerRules'
import { TriggerEvents } from './trigger-tabs/TriggerEvents'
import { TriggerRule } from './TriggerRule'
import { TriggerEvent } from './TriggerEvent'
import { NotFound } from '@/views/NotFound'
import { appRoutes } from '@/views/app/routes'
import { installRoutes } from '@/views/install/routes'
import { useOrg } from '@/hooks/use-org'

const TriggersGate = () => {
  const { org } = useOrg()

  if (!org) return null
  if (!org?.features?.['triggers']) return <NotFound />
  return <Outlet />
}

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
      {
        path: ':orgId/runner/processes/:processId/logs',
        element: <ProcessSystemLogs />,
      },
      { path: ':orgId/team', element: <Team /> },
      { path: ':orgId/api-tokens', element: <ApiTokens /> },
      { path: ':orgId/service-accounts', element: <ServiceAccounts /> },
      { path: ':orgId/webhooks', element: <Webhooks /> },
      {
        element: <TriggersGate />,
        children: [
          { path: ':orgId/triggers', element: <Triggers /> },
          {
            path: ':orgId/triggers/:triggerId',
            element: <TriggerLayout />,
            children: [
              { index: true, element: <TriggerOverview /> },
              { path: 'rules', element: <TriggerRules /> },
              { path: 'rules/:ruleId', element: <TriggerRule /> },
              { path: 'events', element: <TriggerEvents /> },
              { path: 'events/:eventId', element: <TriggerEvent /> },
            ],
          },
        ],
      },
      { path: ':orgId/slack', element: <Slack /> },
      {
        path: ':orgId/connections',
        loader: ({ params }) => redirect(`/${params.orgId}`),
      },
      {
        path: ':orgId/connections/vcs',
        loader: ({ params }) => redirect(`/${params.orgId}`),
      },
      {
        path: ':orgId/connections/vcs/:connectionId',
        element: <VCSConnectionDetail />,
      },
      {
        path: ':orgId/connections/:connectionId',
        loader: ({ params }) =>
          redirect(`/${params.orgId}/connections/vcs/${params.connectionId}`),
      },
      ...appRoutes,
      ...installRoutes,
      { path: ':orgId/*', element: <NotFound /> },
    ],
  },
]
