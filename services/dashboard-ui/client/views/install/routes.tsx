import { Outlet, redirect, useMatch, type RouteObject } from 'react-router'
import { useNewInstallIA } from '@/hooks/use-new-install-ia'
import { useOrg } from '@/hooks/use-org'
import { NotFound } from '@/views/NotFound'
import { InstallLayout } from './InstallLayout'
import { Overview } from './Overview'
import { Components } from './Components'
import { Resources } from './Resources'
import {
  NewInstallPlaceholder,
  NewInstallPlaceholderBody,
} from './NewInstallPlaceholder'
import { Deployments } from './Deployments'
import {
  NewInstallConfigurationLayout,
  NewInstallOperationsLayout,
  NewInstallResourcesLayout,
} from './NewInstallSectionLayout'
import { NewInstallComponents } from './NewInstallComponents'
import { NewInstallImages } from './NewInstallImages'
import { NewInstallSandbox } from './NewInstallSandbox'
import { NewInstallHealth } from './NewInstallHealth'
import { NewInstallState } from './NewInstallState'
import { NewInstallStack } from './NewInstallStack'
import {
  NewInstallAppBranch,
  NewInstallConfigFile,
  NewInstallInputs,
  NewInstallOverrides,
} from './NewInstallConfiguration'
import { Actions } from './Actions'
import { Roles } from './Roles'
import { Policies } from './Policies'
import { Runner } from './Runner'
import { ProcessSystemLogs } from './ProcessSystemLogs'
import { Sandbox } from './Sandbox'
import { Stacks } from './Stacks'
import { Updates } from './Updates'
import { History } from './History'
import { Readme } from './Readme'
import { InstallComponentLayout } from './InstallComponentLayout'
import { InstallComponentOverviewTab } from './install-component-tabs/InstallComponentOverviewTab'
import { InstallComponentDeploysTab } from './install-component-tabs/InstallComponentDeploysTab'
import { InstallComponentConfigTab } from './install-component-tabs/InstallComponentConfigTab'
import { InstallComponentStateTab } from './install-component-tabs/InstallComponentStateTab'
import { DeployLayout } from './DeployLayout'
import { DeploySummaryTab } from './deploy-tabs/DeploySummaryTab'
import { DeployLogsTab } from './deploy-tabs/DeployLogsTab'
import { DeployTraceTab } from './deploy-tabs/DeployTraceTab'
import { DeployPlanTab } from './deploy-tabs/DeployPlanTab'
import { DeployVariablesTab } from './deploy-tabs/DeployVariablesTab'
import { DeployStateTab } from './deploy-tabs/DeployStateTab'
import { DeployValuesTab } from './deploy-tabs/DeployValuesTab'
import { DeployOutputsTab } from './deploy-tabs/DeployOutputsTab'
import { DeployManifestTab } from './deploy-tabs/DeployManifestTab'
import { DeployArtifactTab } from './deploy-tabs/DeployArtifactTab'
import { ActionDetail } from './ActionDetail'
import { Runbooks } from './Runbooks'
import { RunbookDetailLayout } from './RunbookDetailLayout'
import { RunbookReadmeTab } from './runbook-tabs/RunbookReadmeTab'
import { RunbookStepsTab } from './runbook-tabs/RunbookStepsTab'
import { RunbookHistoryTab } from './runbook-tabs/RunbookHistoryTab'
import { ActionRunLayout } from './ActionRunLayout'
import { ActionRunDetail } from './ActionRunDetail'
import { ActionRunLogsPage } from './ActionRunLogs'
import { ActionRunTracePage } from './ActionRunTrace'
import { SandboxRunLayout } from './SandboxRunLayout'
import { SandboxRunSummaryTab } from './sandbox-tabs/SandboxRunSummaryTab'
import { SandboxRunLogsTab } from './sandbox-tabs/SandboxRunLogsTab'
import { SandboxRunPlanTab } from './sandbox-tabs/SandboxRunPlanTab'
import { SandboxRunTraceTab } from './sandbox-tabs/SandboxRunTraceTab'
import { SandboxRunVariablesTab } from './sandbox-tabs/SandboxRunVariablesTab'
import { SandboxRunStateTab } from './sandbox-tabs/SandboxRunStateTab'
import { SandboxRunOutputsTab } from './sandbox-tabs/SandboxRunOutputsTab'
import { CurrentInputs } from './CurrentInputs'
import { ViewState } from './ViewState'
import { WorkflowDetail } from './WorkflowDetail'
import { RunnerJobDetail } from './RunnerJobDetail'
import { Notebooks } from './Notebooks'
import { NotebookDetail } from './NotebookDetail'
import { InstallConfigs } from './InstallConfigs'

const legacyRedirect =
  (to: (params: Record<string, string | undefined>) => string) =>
  ({
    params,
    request,
  }: {
    params: Record<string, string | undefined>
    request: Request
  }) => {
    const { search, hash } = new URL(request.url)
    return redirect(`${to(params)}${search}${hash}`)
  }

const NewInstallIAGate = () => {
  const { org } = useOrg()
  const hasNewInstallIA = useNewInstallIA()

  if (!org) return null
  return hasNewInstallIA ? <Outlet /> : <NotFound />
}

const InstallOverviewRoute = () => {
  const hasNewInstallIA = useNewInstallIA()

  return hasNewInstallIA ? (
    <NewInstallPlaceholder path="" title="Overview" />
  ) : (
    <Overview />
  )
}

const InstallResourcesRoute = () => {
  const hasNewInstallIA = useNewInstallIA()
  const isIndex = !!useMatch('/:orgId/installs/:installId/resources')

  if (hasNewInstallIA) return <NewInstallResourcesLayout />
  if (isIndex) return <Resources />
  return <Outlet />
}

const NewInstallResourcesIndex = () => {
  const hasNewInstallIA = useNewInstallIA()
  if (!hasNewInstallIA) return null
  return <NewInstallStack />
}

export const installRoutes: RouteObject[] = [
  {
    element: <InstallLayout />,
    children: [
      {
        path: ':orgId/installs/:installId',
        element: <InstallOverviewRoute />,
      },
      {
        path: ':orgId/installs/:installId/components',
        element: <Components />,
      },
      {
        path: ':orgId/installs/:installId/resources',
        element: <InstallResourcesRoute />,
        children: [
          { index: true, element: <NewInstallResourcesIndex /> },
          {
            element: <NewInstallIAGate />,
            children: [
              {
                path: 'sandbox',
                element: <NewInstallSandbox />,
              },
              {
                path: 'components',
                element: <NewInstallComponents />,
              },
              {
                path: 'images',
                element: <NewInstallImages />,
              },
            ],
          },
        ],
      },
      {
        element: <NewInstallIAGate />,
        children: [
          {
            path: ':orgId/installs/:installId/deployments',
            element: <Deployments />,
          },
          {
            path: ':orgId/installs/:installId/health',
            element: <NewInstallHealth />,
          },
          {
            path: ':orgId/installs/:installId/operations',
            element: <NewInstallOperationsLayout />,
            children: [
              {
                index: true,
                element: <NewInstallPlaceholderBody title="Activity" />,
              },
              {
                path: 'actions',
                element: <NewInstallPlaceholderBody title="Actions" />,
              },
              {
                path: 'runbooks',
                element: <NewInstallPlaceholderBody title="Runbooks" />,
              },
              {
                path: 'policies',
                element: <NewInstallPlaceholderBody title="Policies" />,
              },
              {
                path: 'runner',
                element: <NewInstallPlaceholderBody title="Runner" />,
              },
            ],
          },
          {
            path: ':orgId/installs/:installId/configuration',
            element: <NewInstallConfigurationLayout />,
            children: [
              {
                index: true,
                element: <NewInstallAppBranch />,
              },
              {
                path: 'inputs',
                element: <NewInstallInputs />,
              },
              {
                path: 'config-file',
                element: <NewInstallConfigFile />,
              },
              {
                path: 'overrides',
                element: <NewInstallOverrides />,
              },
              {
                path: 'state',
                element: <NewInstallState />,
              },
            ],
          },
        ],
      },
      { path: ':orgId/installs/:installId/actions', element: <Actions /> },
      { path: ':orgId/installs/:installId/notebooks', element: <Notebooks /> },
      {
        path: ':orgId/installs/:installId/notebooks/:notebookId',
        element: <NotebookDetail />,
      },
      { path: ':orgId/installs/:installId/roles', element: <Roles /> },
      { path: ':orgId/installs/:installId/policies', element: <Policies /> },
      { path: ':orgId/installs/:installId/runner', element: <Runner /> },
      { path: ':orgId/installs/:installId/inputs', element: <CurrentInputs /> },
      { path: ':orgId/installs/:installId/state', element: <ViewState /> },
      {
        path: ':orgId/installs/:installId/runner/jobs/:jobId',
        element: <RunnerJobDetail />,
      },
      {
        path: ':orgId/installs/:installId/runner/processes/:processId/logs',
        element: <ProcessSystemLogs />,
      },
      { path: ':orgId/installs/:installId/sandbox', element: <Sandbox /> },
      {
        path: ':orgId/installs/:installId/sandbox/runs',
        loader: ({ params }) =>
          redirect(`/${params.orgId}/installs/${params.installId}/sandbox`),
      },
      {
        path: ':orgId/installs/:installId/sandbox/runs/:runId',
        element: <SandboxRunLayout />,
        children: [
          { index: true, element: <SandboxRunSummaryTab /> },
          { path: 'logs', element: <SandboxRunLogsTab /> },
          { path: 'trace', element: <SandboxRunTraceTab /> },
          { path: 'plan', element: <SandboxRunPlanTab /> },
          { path: 'variables', element: <SandboxRunVariablesTab /> },
          { path: 'state', element: <SandboxRunStateTab /> },
          { path: 'outputs', element: <SandboxRunOutputsTab /> },
        ],
      },
      {
        path: ':orgId/installs/:installId/history/:workflowId',
        element: <WorkflowDetail />,
      },
      {
        path: ':orgId/installs/:installId/workflows/:workflowId',
        loader: legacyRedirect(
          (params) =>
            `/${params.orgId}/installs/${params.installId}/history/${params.workflowId}`
        ),
      },
      { path: ':orgId/installs/:installId/stacks', element: <Stacks /> },
      {
        path: ':orgId/installs/:installId/updates',
        element: <Updates />,
      },
      {
        path: ':orgId/installs/:installId/history',
        element: <History />,
      },
      {
        path: ':orgId/installs/:installId/app-branch-runs',
        loader: legacyRedirect(
          (params) => `/${params.orgId}/installs/${params.installId}/updates`
        ),
      },
      {
        path: ':orgId/installs/:installId/versions',
        loader: legacyRedirect(
          (params) => `/${params.orgId}/installs/${params.installId}/updates`
        ),
      },
      {
        path: ':orgId/installs/:installId/workflows',
        loader: legacyRedirect(
          (params) => `/${params.orgId}/installs/${params.installId}/history`
        ),
      },
      {
        path: ':orgId/installs/:installId/configs',
        element: <InstallConfigs />,
      },
      { path: ':orgId/installs/:installId/readme', element: <Readme /> },
      {
        path: ':orgId/installs/:installId/components/:componentId',
        element: <InstallComponentLayout />,
        children: [
          { index: true, element: <InstallComponentOverviewTab /> },
          { path: 'deploys', element: <InstallComponentDeploysTab /> },
          { path: 'config', element: <InstallComponentConfigTab /> },
          { path: 'state', element: <InstallComponentStateTab /> },
        ],
      },
      {
        path: ':orgId/installs/:installId/components/:componentId/deploys/:deployId',
        element: <DeployLayout />,
        children: [
          { index: true, element: <DeploySummaryTab /> },
          { path: 'logs', element: <DeployLogsTab /> },
          { path: 'trace', element: <DeployTraceTab /> },
          { path: 'plan', element: <DeployPlanTab /> },
          { path: 'variables', element: <DeployVariablesTab /> },
          { path: 'state', element: <DeployStateTab /> },
          { path: 'values', element: <DeployValuesTab /> },
          { path: 'outputs', element: <DeployOutputsTab /> },
          { path: 'manifest', element: <DeployManifestTab /> },
          { path: 'artifact', element: <DeployArtifactTab /> },
        ],
      },
      {
        path: ':orgId/installs/:installId/actions/:actionId',
        element: <ActionDetail />,
      },
      { path: ':orgId/installs/:installId/runbooks', element: <Runbooks /> },
      {
        path: ':orgId/installs/:installId/runbooks/:runbookId',
        element: <RunbookDetailLayout />,
        children: [
          { path: 'readme', element: <RunbookReadmeTab /> },
          { path: 'steps', element: <RunbookStepsTab /> },
          { path: 'history', element: <RunbookHistoryTab /> },
        ],
      },
      {
        path: ':orgId/installs/:installId/actions/:actionId/runs',
        loader: ({ params }) =>
          redirect(
            `/${params.orgId}/installs/${params.installId}/actions/${params.actionId}`
          ),
      },
      {
        path: ':orgId/installs/:installId/actions/:actionId/runs/:actionRunId',
        element: <ActionRunLayout />,
        children: [
          { index: true, element: <ActionRunDetail /> },
          { path: 'logs', element: <ActionRunLogsPage /> },
          { path: 'trace', element: <ActionRunTracePage /> },
        ],
      },
    ],
  },
]
