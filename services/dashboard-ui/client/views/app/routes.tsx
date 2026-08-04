import { lazy, Suspense, type ReactNode } from 'react'
import { redirect, type RouteObject, type Params } from 'react-router'
import { AppLayout } from './AppLayout'
import { AppIndex } from './AppIndex'
import { LegacyAppRoute } from './LegacyAppRoute'
import { Components } from './Components'
import { ComponentDetail } from './ComponentDetail'
import { BuildDetail } from './BuildDetail'
import { Actions } from './Actions'
import { ActionDetail } from './ActionDetail'
import { Runbooks } from './Runbooks'
import { OperationsStudio } from './OperationsStudio'
import { RunbookDetailLayout } from './RunbookDetailLayout'
import { RunbookReadmeTab } from './runbook-tabs/RunbookReadmeTab'
import { RunbookStepsTab } from './runbook-tabs/RunbookStepsTab'
import { Roles } from './Roles'
import { PoliciesLayout } from './PoliciesLayout'
import { Policies } from './Policies'
import { PolicyDetail } from './PolicyDetail'

const PolicyAnalytics = lazy(() =>
  import('./PolicyAnalytics').then((m) => ({ default: m.PolicyAnalytics }))
)
import { Installs } from './Installs'
import { Labels } from './Labels'
import { Readme } from './Readme'
import { Sandbox } from './Sandbox'
import { SandboxBuildDetail } from './SandboxBuildDetail'
import { Branches } from './branches/Branches'
import { BranchLayout } from './branches/BranchLayout'
import { BranchOverviewTab } from './branches/tabs/BranchOverviewTab'
import { BranchRunsTab } from './branches/tabs/BranchRunsTab'
import { BranchPlanTab } from './branches/tabs/BranchPlanTab'
import { BranchConfigsTab } from './branches/tabs/BranchConfigsTab'
import { BranchComponents } from './branches/scoped/BranchComponents'
import { BranchActions } from './branches/scoped/BranchActions'
import { BranchRunbooks } from './branches/scoped/BranchRunbooks'
import { BranchInstalls } from './branches/scoped/BranchInstalls'
import { BranchRunDetail } from './branches/BranchRunDetail'

const legacy = (
  element: ReactNode,
  subPath?: (params: Params) => string
) => <LegacyAppRoute subPath={subPath}>{element}</LegacyAppRoute>

export const appRoutes: RouteObject[] = [
  {
    element: <AppLayout />,
    children: [
      { path: ':orgId/apps/:appId', element: <AppIndex /> },
      {
        path: ':orgId/apps/:appId/components',
        element: legacy(<Components />, () => 'components'),
      },
      {
        path: ':orgId/apps/:appId/components/:componentId',
        element: legacy(
          <ComponentDetail />,
          (p) => `components/${p.componentId}`
        ),
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
        element: legacy(
          <BuildDetail />,
          (p) => `components/${p.componentId}/builds/${p.buildId}`
        ),
      },
      {
        path: ':orgId/apps/:appId/actions',
        element: legacy(<Actions />, () => 'actions'),
      },
      {
        path: ':orgId/apps/:appId/actions/:actionId',
        element: legacy(<ActionDetail />, (p) => `actions/${p.actionId}`),
      },
      {
        path: ':orgId/apps/:appId/runbooks',
        element: legacy(<Runbooks />, () => 'runbooks'),
      },
      { path: ':orgId/apps/:appId/studio', element: <OperationsStudio /> },
      {
        path: ':orgId/apps/:appId/runbooks/builder',
        loader: ({ params }) =>
          redirect(`/${params.orgId}/apps/${params.appId}/studio`),
      },
      {
        path: ':orgId/apps/:appId/runbooks/:runbookId',
        element: legacy(
          <RunbookDetailLayout />,
          (p) => `runbooks/${p.runbookId}`
        ),
        children: [
          { index: true, element: <RunbookReadmeTab /> },
          { path: 'steps', element: <RunbookStepsTab /> },
        ],
      },
      {
        path: ':orgId/apps/:appId/roles',
        element: legacy(<Roles />, () => 'roles'),
      },
      {
        path: ':orgId/apps/:appId/policies',
        element: legacy(<PoliciesLayout />, () => 'policies'),
        children: [
          { index: true, element: <Policies /> },
          { path: 'analytics', element: <Suspense><PolicyAnalytics /></Suspense> },
        ],
      },
      {
        path: ':orgId/apps/:appId/policies/:policyId',
        element: legacy(<PolicyDetail />, (p) => `policies/${p.policyId}`),
      },
      { path: ':orgId/apps/:appId/branches', element: legacy(<Branches />) },
      {
        path: ':orgId/apps/:appId/branches/:branchId',
        element: <BranchLayout />,
        children: [
          { index: true, element: <BranchOverviewTab /> },
          { path: 'runs', element: <BranchRunsTab /> },
          { path: 'runs/:runId', element: <BranchRunDetail /> },
          { path: 'plan', element: <BranchPlanTab /> },
          { path: 'configs', element: <BranchConfigsTab /> },
          { path: 'components', element: <BranchComponents /> },
          { path: 'components/:componentId', element: <ComponentDetail /> },
          {
            path: 'components/:componentId/builds/:buildId',
            element: <BuildDetail />,
          },
          { path: 'actions', element: <BranchActions /> },
          { path: 'actions/:actionId', element: <ActionDetail /> },
          { path: 'runbooks', element: <BranchRunbooks /> },
          {
            path: 'runbooks/:runbookId',
            element: <RunbookDetailLayout />,
            children: [
              { index: true, element: <RunbookReadmeTab /> },
              { path: 'steps', element: <RunbookStepsTab /> },
            ],
          },
          { path: 'roles', element: <Roles /> },
          {
            path: 'policies',
            element: <PoliciesLayout />,
            children: [
              { index: true, element: <Policies /> },
              { path: 'analytics', element: <Suspense><PolicyAnalytics /></Suspense> },
            ],
          },
          { path: 'policies/:policyId', element: <PolicyDetail /> },
          { path: 'installs', element: <BranchInstalls /> },
          { path: 'labels', element: <Labels /> },
          { path: 'readme', element: <Readme /> },
          { path: 'sandbox', element: <Sandbox /> },
          { path: 'sandbox/builds/:buildId', element: <SandboxBuildDetail /> },
        ],
      },
      {
        path: ':orgId/apps/:appId/sandbox',
        element: legacy(<Sandbox />, () => 'sandbox'),
      },
      {
        path: ':orgId/apps/:appId/sandbox/builds/:buildId',
        element: legacy(
          <SandboxBuildDetail />,
          (p) => `sandbox/builds/${p.buildId}`
        ),
      },
      {
        path: ':orgId/apps/:appId/installs',
        element: legacy(<Installs />, () => 'installs'),
      },
      {
        path: ':orgId/apps/:appId/labels',
        element: legacy(<Labels />, () => 'labels'),
      },
      {
        path: ':orgId/apps/:appId/readme',
        element: legacy(<Readme />, () => 'readme'),
      },
      {
        path: ':orgId/apps/:appId/readme-studio',
        loader: ({ params }) =>
          redirect(`/${params.orgId}/apps/${params.appId}/studio`),
      },
    ],
  },
]
