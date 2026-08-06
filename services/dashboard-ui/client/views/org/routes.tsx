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
import { OIDCTrustPolicies } from './OIDCTrustPolicies'
import { Triggers } from './Triggers'
import { TriggerLayout } from './TriggerLayout'
import { TriggerOverview } from './trigger-tabs/TriggerOverview'
import { TriggerRules } from './trigger-tabs/TriggerRules'
import { TriggerEvents } from './trigger-tabs/TriggerEvents'
import { TriggerRule } from './TriggerRule'
import { TriggerEvent } from './TriggerEvent'
import { SettingsLayout } from '@/views/settings/SettingsLayout'
import { VCSConnections } from '@/views/settings/VCSConnections'
import { NotFound } from '@/views/NotFound'
import { appRoutes } from '@/views/app/routes'
import { installRoutes } from '@/views/install/routes'
import { useOrg } from '@/hooks/use-org'
import { useCLIConfig } from '@/hooks/use-cli-config'

const TriggersGate = () => {
  const { org } = useOrg()

  if (!org) return null
  if (!org?.features?.['triggers']) return <NotFound />
  return <Outlet />
}

const OIDCFederationGate = () => {
  const { data: cliConfig, isLoading } = useCLIConfig()

  if (isLoading) return null
  if (!cliConfig?.oidc_federation_enabled) return <NotFound />
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
      {
        element: <SettingsLayout />,
        children: [
          {
            path: ':orgId/settings',
            loader: ({ params }) => redirect(`/${params.orgId}/settings/vcs`),
          },
          { path: ':orgId/settings/vcs', element: <VCSConnections /> },
          {
            path: ':orgId/settings/vcs/:connectionId',
            element: <VCSConnectionDetail />,
          },
          { path: ':orgId/settings/webhooks', element: <Webhooks /> },
          { path: ':orgId/settings/api-tokens', element: <ApiTokens /> },
          {
            path: ':orgId/settings/service-accounts',
            element: <ServiceAccounts />,
          },
          {
            element: <OIDCFederationGate />,
            children: [
              { path: ':orgId/settings/oidc', element: <OIDCTrustPolicies /> },
            ],
          },
          {
            element: <TriggersGate />,
            children: [
              { path: ':orgId/settings/triggers', element: <Triggers /> },
              {
                path: ':orgId/settings/triggers/:triggerId',
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
          { path: ':orgId/settings/slack', element: <Slack /> },
        ],
      },
      {
        path: ':orgId/webhooks',
        loader: ({ params }) => redirect(`/${params.orgId}/settings/webhooks`),
      },
      {
        path: ':orgId/api-tokens',
        loader: ({ params }) => redirect(`/${params.orgId}/settings/api-tokens`),
      },
      {
        path: ':orgId/service-accounts',
        loader: ({ params }) =>
          redirect(`/${params.orgId}/settings/service-accounts`),
      },
      {
        path: ':orgId/oidc-trust-policies',
        loader: ({ params }) => redirect(`/${params.orgId}/settings/oidc`),
      },
      {
        path: ':orgId/triggers',
        loader: ({ params }) => redirect(`/${params.orgId}/settings/triggers`),
      },
      {
        path: ':orgId/triggers/:triggerId/*',
        loader: ({ params }) =>
          redirect(`/${params.orgId}/settings/triggers/${params.triggerId}`),
      },
      {
        path: ':orgId/slack',
        loader: ({ params }) => redirect(`/${params.orgId}/settings/slack`),
      },
      {
        path: ':orgId/connections',
        loader: ({ params }) => redirect(`/${params.orgId}/settings/vcs`),
      },
      {
        path: ':orgId/connections/vcs',
        loader: ({ params }) => redirect(`/${params.orgId}/settings/vcs`),
      },
      {
        path: ':orgId/connections/vcs/:connectionId',
        loader: ({ params }) =>
          redirect(`/${params.orgId}/settings/vcs/${params.connectionId}`),
      },
      {
        path: ':orgId/connections/:connectionId',
        loader: ({ params }) =>
          redirect(`/${params.orgId}/settings/vcs/${params.connectionId}`),
      },
      ...appRoutes,
      ...installRoutes,
      { path: ':orgId/*', element: <NotFound /> },
    ],
  },
]
