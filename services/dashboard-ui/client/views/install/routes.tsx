import { redirect, type RouteObject } from 'react-router'
import { InstallLayout } from './InstallLayout'
import { Overview } from './Overview'
import { Components } from './Components'
import { Actions } from './Actions'
import { Roles } from './Roles'
import { Policies } from './Policies'
import { Runner } from './Runner'
import { Sandbox } from './Sandbox'
import { Stacks } from './Stacks'
import { Workflows } from './Workflows'
import { Readme } from './Readme'
import { InstallComponentDetail } from './ComponentDetail'
import { DeployDetailGate } from './DeployDetailGate'
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
import { SandboxRunDetailGate } from './SandboxRunDetailGate'
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

export const installRoutes: RouteObject[] = [
  {
    element: <InstallLayout />,
    children: [
      { path: ':orgId/installs/:installId', element: <Overview /> },
      {
        path: ':orgId/installs/:installId/components',
        element: <Components />,
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
      { path: ':orgId/installs/:installId/sandbox', element: <Sandbox /> },
      {
        path: ':orgId/installs/:installId/sandbox/runs',
        loader: ({ params }) =>
          redirect(`/${params.orgId}/installs/${params.installId}/sandbox`),
      },
      {
        path: ':orgId/installs/:installId/sandbox/runs/:runId',
        element: <SandboxRunDetailGate />,
        children: [
          { index: true, element: <SandboxRunLogsTab /> },
          { path: 'trace', element: <SandboxRunTraceTab /> },
          { path: 'plan', element: <SandboxRunPlanTab /> },
          { path: 'variables', element: <SandboxRunVariablesTab /> },
          { path: 'state', element: <SandboxRunStateTab /> },
          { path: 'outputs', element: <SandboxRunOutputsTab /> },
        ],
      },
      {
        path: ':orgId/installs/:installId/workflows/:workflowId',
        element: <WorkflowDetail />,
      },
      { path: ':orgId/installs/:installId/stacks', element: <Stacks /> },
      { path: ':orgId/installs/:installId/workflows', element: <Workflows /> },
      { path: ':orgId/installs/:installId/readme', element: <Readme /> },
      {
        path: ':orgId/installs/:installId/components/:componentId',
        element: <InstallComponentDetail />,
      },
      {
        path: ':orgId/installs/:installId/components/:componentId/deploys',
        loader: ({ params }) =>
          redirect(
            `/${params.orgId}/installs/${params.installId}/components/${params.componentId}`
          ),
      },
      {
        path: ':orgId/installs/:installId/components/:componentId/deploys/:deployId',
        element: <DeployDetailGate />,
        children: [
          { index: true, element: <DeployLogsTab /> },
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
