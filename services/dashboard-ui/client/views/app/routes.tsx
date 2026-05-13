import { lazy, Suspense } from 'react'
import { redirect, type RouteObject } from 'react-router'
import { AppLayout } from './AppLayout'
import { Overview } from './Overview'
import { Components } from './Components'
import { ComponentDetail } from './ComponentDetail'
import { BuildDetail } from './BuildDetail'
import { Actions } from './Actions'
import { ActionDetail } from './ActionDetail'
import { Roles } from './Roles'
import { PoliciesLayout } from './PoliciesLayout'
import { Policies } from './Policies'
import { PolicyDetail } from './PolicyDetail'

const PolicyAnalytics = lazy(() =>
  import('./PolicyAnalytics').then((m) => ({ default: m.PolicyAnalytics }))
)
import { Installs } from './Installs'
import { Readme } from './Readme'
import { Sandbox } from './Sandbox'
import { SandboxBuildDetail } from './SandboxBuildDetail'
import { Branches } from './branches/Branches'
import { BranchDetail } from './branches/BranchDetail'
import { BranchConfigs } from './branches/BranchConfigs'
import { BranchRunDetail } from './branches/BranchRunDetail'

export const appRoutes: RouteObject[] = [
  {
    element: <AppLayout />,
    children: [
      { path: ':orgId/apps/:appId', element: <Overview /> },
      { path: ':orgId/apps/:appId/components', element: <Components /> },
      {
        path: ':orgId/apps/:appId/components/:componentId',
        element: <ComponentDetail />,
      },
      {
        path: ':orgId/apps/:appId/components/:componentId/builds',
        loader: ({ params }) =>
          redirect(
            `/${params.orgId}/apps/${params.appId}/components/${params.componentId}`
          ),
      },
      {
        path: ':orgId/apps/:appId/components/:componentId/builds/:buildId',
        element: <BuildDetail />,
      },
      { path: ':orgId/apps/:appId/actions', element: <Actions /> },
      {
        path: ':orgId/apps/:appId/actions/:actionId',
        element: <ActionDetail />,
      },
      { path: ':orgId/apps/:appId/roles', element: <Roles /> },
      {
        path: ':orgId/apps/:appId/policies',
        element: <PoliciesLayout />,
        children: [
          { index: true, element: <Policies /> },
          { path: 'analytics', element: <Suspense><PolicyAnalytics /></Suspense> },
        ],
      },
      {
        path: ':orgId/apps/:appId/policies/:policyId',
        element: <PolicyDetail />,
      },
      { path: ':orgId/apps/:appId/branches', element: <Branches /> },
      {
        path: ':orgId/apps/:appId/branches/:branchId',
        element: <BranchDetail />,
      },
      {
        path: ':orgId/apps/:appId/branches/:branchId/configs',
        element: <BranchConfigs />,
      },
      {
        path: ':orgId/apps/:appId/branches/:branchId/runs',
        loader: ({ params }) =>
          redirect(
            `/${params.orgId}/apps/${params.appId}/branches/${params.branchId}`
          ),
      },
      {
        path: ':orgId/apps/:appId/branches/:branchId/runs/:runId',
        element: <BranchRunDetail />,
      },
      { path: ':orgId/apps/:appId/sandbox', element: <Sandbox /> },
      {
        path: ':orgId/apps/:appId/sandbox/builds/:buildId',
        element: <SandboxBuildDetail />,
      },
      { path: ':orgId/apps/:appId/installs', element: <Installs /> },
      { path: ':orgId/apps/:appId/readme', element: <Readme /> },
    ],
  },
]
