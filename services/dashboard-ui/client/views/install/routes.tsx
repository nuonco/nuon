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
import { DeployDetail } from './DeployDetail'
import { ActionDetail } from './ActionDetail'
import { ActionRunLayout } from './ActionRunLayout'
import { ActionRunDetail } from './ActionRunDetail'
import { ActionRunLogsPage } from './ActionRunLogs'
import { SandboxRunDetail } from './SandboxRunDetail'
import { WorkflowDetail } from './WorkflowDetail'

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
      { path: ':orgId/installs/:installId/roles', element: <Roles /> },
      { path: ':orgId/installs/:installId/policies', element: <Policies /> },
      { path: ':orgId/installs/:installId/runner', element: <Runner /> },
      { path: ':orgId/installs/:installId/sandbox', element: <Sandbox /> },
      {
        path: ':orgId/installs/:installId/sandbox/runs',
        loader: ({ params }) =>
          redirect(`/${params.orgId}/installs/${params.installId}/sandbox`),
      },
      {
        path: ':orgId/installs/:installId/sandbox/runs/:runId',
        element: <SandboxRunDetail />,
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
        element: <DeployDetail />,
      },
      {
        path: ':orgId/installs/:installId/actions/:actionId',
        element: <ActionDetail />,
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
        ],
      },
    ],
  },
]
