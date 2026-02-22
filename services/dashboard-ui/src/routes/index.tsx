import React, { lazy, Suspense } from 'react'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'

function LoadingSpinner() {
  return (
    <div className="flex items-center justify-center h-screen">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900" />
    </div>
  )
}

function NotFound() {
  return (
    <div className="flex items-center justify-center h-screen">
      <div className="text-center">
        <h1 className="text-2xl font-bold mb-4">Page Not Found</h1>
        <p className="text-gray-600 mb-4">The page you are looking for does not exist.</p>
        <a href="/" className="text-blue-600 hover:underline">Go to Home</a>
      </div>
    </div>
  )
}

const HomePage = lazy(() => import('@/pages/HomePage'))

const OrgLayout = lazy(() => import('@/pages/layouts/OrgLayout'))
const AppLayout = lazy(() => import('@/pages/layouts/AppLayout'))
const InstallLayout = lazy(() => import('@/pages/layouts/InstallLayout'))
const InstallActionRunLayout = lazy(() => import('@/pages/layouts/InstallActionRunLayout'))

const OrgDashboard = lazy(() => import('@/pages/org/OrgDashboard'))
const AppsPage = lazy(() => import('@/pages/org/AppsPage'))
const InstallsPage = lazy(() => import('@/pages/org/InstallsPage'))
const OrgRunner = lazy(() => import('@/pages/org/OrgRunner'))
const TeamPage = lazy(() => import('@/pages/org/TeamPage'))

const AppOverview = lazy(() => import('@/pages/apps/AppOverview'))
const AppComponents = lazy(() => import('@/pages/apps/AppComponents'))
const AppComponentDetail = lazy(() => import('@/pages/apps/AppComponentDetail'))
const AppInstalls = lazy(() => import('@/pages/apps/AppInstalls'))
const AppActions = lazy(() => import('@/pages/apps/AppActions'))
const AppActionDetail = lazy(() => import('@/pages/apps/AppActionDetail'))
const AppPolicies = lazy(() => import('@/pages/apps/AppPolicies'))
const AppPolicyDetail = lazy(() => import('@/pages/apps/AppPolicyDetail'))
const AppReadme = lazy(() => import('@/pages/apps/AppReadme'))
const AppRoles = lazy(() => import('@/pages/apps/AppRoles'))

const InstallOverview = lazy(() => import('@/pages/installs/InstallOverview'))
const InstallComponents = lazy(() => import('@/pages/installs/InstallComponents'))
const InstallComponentDetail = lazy(() => import('@/pages/installs/InstallComponentDetail'))
const InstallWorkflows = lazy(() => import('@/pages/installs/InstallWorkflows'))
const InstallWorkflowDetail = lazy(() => import('@/pages/installs/InstallWorkflowDetail'))
const InstallActions = lazy(() => import('@/pages/installs/InstallActions'))
const InstallActionDetail = lazy(() => import('@/pages/installs/InstallActionDetail'))
const InstallActionRunSummary = lazy(() => import('@/pages/installs/InstallActionRunSummary'))
const InstallActionRunLogs = lazy(() => import('@/pages/installs/InstallActionRunLogs'))
const InstallRunner = lazy(() => import('@/pages/installs/InstallRunner'))
const InstallSandbox = lazy(() => import('@/pages/installs/InstallSandbox'))
const InstallSandboxRun = lazy(() => import('@/pages/installs/InstallSandboxRun'))
const InstallPolicies = lazy(() => import('@/pages/installs/InstallPolicies'))
const InstallRoles = lazy(() => import('@/pages/installs/InstallRoles'))
const InstallStacks = lazy(() => import('@/pages/installs/InstallStacks'))

function wrap(Component: React.ComponentType) {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <Component />
    </Suspense>
  )
}

const router = createBrowserRouter([
  {
    path: '/',
    element: wrap(HomePage),
  },
  {
    path: '/:orgId',
    element: wrap(OrgLayout),
    children: [
      {
        index: true,
        element: wrap(OrgDashboard),
      },
      {
        path: 'apps',
        element: wrap(AppsPage),
      },
      {
        path: 'apps/:appId',
        element: wrap(AppLayout),
        children: [
          {
            index: true,
            element: wrap(AppOverview),
          },
          {
            path: 'components',
            element: wrap(AppComponents),
          },
          {
            path: 'components/:componentId',
            element: wrap(AppComponentDetail),
          },
          {
            path: 'installs',
            element: wrap(AppInstalls),
          },
          {
            path: 'actions',
            element: wrap(AppActions),
          },
          {
            path: 'actions/:actionId',
            element: wrap(AppActionDetail),
          },
          {
            path: 'policies',
            element: wrap(AppPolicies),
          },
          {
            path: 'policies/:policyId',
            element: wrap(AppPolicyDetail),
          },
          {
            path: 'readme',
            element: wrap(AppReadme),
          },
          {
            path: 'roles',
            element: wrap(AppRoles),
          },
        ],
      },
      {
        path: 'installs',
        element: wrap(InstallsPage),
      },
      {
        path: 'installs/:installId',
        element: wrap(InstallLayout),
        children: [
          {
            index: true,
            element: wrap(InstallOverview),
          },
          {
            path: 'components',
            element: wrap(InstallComponents),
          },
          {
            path: 'components/:componentId',
            element: wrap(InstallComponentDetail),
          },
          {
            path: 'workflows',
            element: wrap(InstallWorkflows),
          },
          {
            path: 'workflows/:workflowId',
            element: wrap(InstallWorkflowDetail),
          },
          {
            path: 'actions',
            element: wrap(InstallActions),
          },
          {
            path: 'actions/:actionId',
            element: wrap(InstallActionDetail),
          },
          {
            path: 'actions/:actionId/:runId',
            element: wrap(InstallActionRunLayout),
            children: [
              {
                index: true,
                element: wrap(InstallActionRunSummary),
              },
              {
                path: 'logs',
                element: wrap(InstallActionRunLogs),
              },
            ],
          },
          {
            path: 'runner',
            element: wrap(InstallRunner),
          },
          {
            path: 'sandbox',
            element: wrap(InstallSandbox),
          },
          {
            path: 'sandbox/:runId',
            element: wrap(InstallSandboxRun),
          },
          {
            path: 'policies',
            element: wrap(InstallPolicies),
          },
          {
            path: 'roles',
            element: wrap(InstallRoles),
          },
          {
            path: 'stacks',
            element: wrap(InstallStacks),
          },
        ],
      },
      {
        path: 'runner',
        element: wrap(OrgRunner),
      },
      {
        path: 'team',
        element: wrap(TeamPage),
      },
    ],
  },
  {
    path: '*',
    element: <NotFound />,
  },
])

export function AppRouter() {
  return <RouterProvider router={router} />
}