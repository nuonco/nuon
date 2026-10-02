import { Outlet, type LoaderFunctionArgs, type RouteObject } from 'react-router'
import { redirect } from 'react-router'
import { useInstallRouteMatch } from '@/hooks/use-install-path'
import { useNewInstallIA } from '@/hooks/use-new-install-ia'
import {
  installRouteShouldRevalidate,
  redirectLegacyInstallRoute,
  redirectNestedInstallRoute,
  redirectWorkflowDetailRoute,
  resolveInstallPrefix,
} from '@/lib/install-routing'
import { useOrg } from '@/hooks/use-org'
import { NotFound } from '@/views/NotFound'
import { InstallLayout } from './InstallLayout'
import { Overview } from './Overview'
import { Components } from './Components'
import { Resources } from './Resources'
import { NewInstallOverview } from './NewInstallOverview'
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
import {
  NewInstallActions,
  NewInstallActivity,
  NewInstallPolicies,
  NewInstallRunbooks,
  NewInstallRunner,
} from './NewInstallOperations'
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
import { DeploymentDetailLayout } from './DeploymentDetailLayout'
import { DeploymentChangesTab } from './deployment-tabs/DeploymentChangesTab'
import { DeploymentTemplateTab } from './deployment-tabs/DeploymentTemplateTab'
import { DeploymentWorkflowTab } from './deployment-tabs/DeploymentWorkflowTab'
import { RunnerJobDetail } from './RunnerJobDetail'
import { Notebooks } from './Notebooks'
import { NotebookDetail } from './NotebookDetail'
import { InstallConfigs } from './InstallConfigs'

const redirectTo =
  (suffix: (params: LoaderFunctionArgs['params']) => string) =>
  async ({ params, request }: LoaderFunctionArgs) => {
    const { search, hash } = new URL(request.url)
    const prefix = await resolveInstallPrefix(params)
    return redirect(`${prefix}${suffix(params)}${search}${hash}`)
  }

const NewInstallIAGate = () => {
  const { org } = useOrg()
  const hasNewInstallIA = useNewInstallIA()

  if (!org) return null
  return hasNewInstallIA ? <Outlet /> : <NotFound />
}

const InstallOverviewRoute = () => {
  const hasNewInstallIA = useNewInstallIA()

  return hasNewInstallIA ? <NewInstallOverview /> : <Overview />
}

const InstallResourcesRoute = () => {
  const hasNewInstallIA = useNewInstallIA()
  const isIndex = !!useInstallRouteMatch('/resources')

  if (hasNewInstallIA) return <NewInstallResourcesLayout />
  if (isIndex) return <Resources />
  return <Outlet />
}

const NewInstallResourcesIndex = () => {
  const hasNewInstallIA = useNewInstallIA()
  if (!hasNewInstallIA) return null
  return <NewInstallStack />
}

const installChildren = (): RouteObject[] => [
  {
    index: true,
    element: <InstallOverviewRoute />,
  },
  {
    path: 'components',
    element: <Components />,
  },
  {
    path: 'resources',
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
        path: 'deployments',
        element: <Deployments />,
      },
      {
        path: 'deployments/:workflowId',
        loader: redirectWorkflowDetailRoute,
        element: <DeploymentDetailLayout />,
        children: [
          { index: true, element: <DeploymentChangesTab /> },
          {
            path: 'template-updates',
            element: <DeploymentTemplateTab />,
          },
          { path: 'workflow', element: <DeploymentWorkflowTab /> },
        ],
      },
      {
        path: 'health',
        element: <NewInstallHealth />,
      },
      {
        path: 'operations',
        element: <NewInstallOperationsLayout />,
        children: [
          {
            index: true,
            element: <NewInstallActivity />,
          },
          {
            path: 'actions',
            element: <NewInstallActions />,
          },
          {
            path: 'runbooks',
            element: <NewInstallRunbooks />,
          },
          {
            path: 'policies',
            element: <NewInstallPolicies />,
          },
          {
            path: 'runner',
            element: <NewInstallRunner />,
          },
        ],
      },
      {
        path: 'configuration',
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
  { path: 'actions', element: <Actions /> },
  { path: 'notebooks', element: <Notebooks /> },
  {
    path: 'notebooks/:notebookId',
    element: <NotebookDetail />,
  },
  { path: 'roles', element: <Roles /> },
  { path: 'policies', element: <Policies /> },
  { path: 'runner', element: <Runner /> },
  { path: 'inputs', element: <CurrentInputs /> },
  { path: 'state', element: <ViewState /> },
  {
    path: 'runner/jobs/:jobId',
    element: <RunnerJobDetail />,
  },
  {
    path: 'runner/processes/:processId/logs',
    element: <ProcessSystemLogs />,
  },
  { path: 'sandbox', element: <Sandbox /> },
  {
    path: 'sandbox/runs',
    loader: redirectTo(() => '/sandbox'),
  },
  {
    path: 'sandbox/runs/:runId',
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
    path: 'history/:workflowId',
    loader: redirectWorkflowDetailRoute,
    element: <WorkflowDetail />,
  },
  {
    path: 'workflows/:workflowId',
    loader: redirectWorkflowDetailRoute,
  },
  { path: 'stacks', element: <Stacks /> },
  {
    path: 'updates',
    element: <Updates />,
  },
  {
    path: 'history',
    element: <History />,
  },
  {
    path: 'app-branch-runs',
    loader: redirectTo(() => '/updates'),
  },
  {
    path: 'versions',
    loader: redirectTo(() => '/updates'),
  },
  {
    path: 'workflows',
    loader: redirectTo(() => '/history'),
  },
  {
    path: 'configs',
    element: <InstallConfigs />,
  },
  { path: 'readme', element: <Readme /> },
  {
    path: 'components/:componentId',
    element: <InstallComponentLayout />,
    children: [
      { index: true, element: <InstallComponentOverviewTab /> },
      { path: 'deploys', element: <InstallComponentDeploysTab /> },
      { path: 'config', element: <InstallComponentConfigTab /> },
      { path: 'state', element: <InstallComponentStateTab /> },
    ],
  },
  {
    path: 'components/:componentId/deploys/:deployId',
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
    path: 'actions/:actionId',
    element: <ActionDetail />,
  },
  { path: 'runbooks', element: <Runbooks /> },
  {
    path: 'runbooks/:runbookId',
    element: <RunbookDetailLayout />,
    children: [
      { path: 'readme', element: <RunbookReadmeTab /> },
      { path: 'steps', element: <RunbookStepsTab /> },
      { path: 'history', element: <RunbookHistoryTab /> },
    ],
  },
  {
    path: 'actions/:actionId/runs',
    loader: redirectTo((params) => `/actions/${params.actionId}`),
  },
  {
    path: 'actions/:actionId/runs/:actionRunId',
    element: <ActionRunLayout />,
    children: [
      { index: true, element: <ActionRunDetail /> },
      { path: 'logs', element: <ActionRunLogsPage /> },
      { path: 'trace', element: <ActionRunTracePage /> },
    ],
  },
]

const installLayoutRoute = (
  path: string,
  loader: RouteObject['loader']
): RouteObject => ({
  path,
  loader,
  shouldRevalidate: installRouteShouldRevalidate,
  element: <InstallLayout />,
  children: installChildren(),
})

export const installRoutes: RouteObject[] = [
  installLayoutRoute(':orgId/installs/:installId', redirectLegacyInstallRoute),
  installLayoutRoute(
    ':orgId/apps/:appId/installs/:installId',
    redirectNestedInstallRoute
  ),
]
