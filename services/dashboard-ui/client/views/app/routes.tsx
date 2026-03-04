import { redirect, type RouteObject } from 'react-router'
import { AppLayout } from './AppLayout'
import { Overview } from './Overview'
import { Components } from './Components'
import { ComponentDetail } from './ComponentDetail'
import { BuildDetail } from './BuildDetail'
import { Actions } from './Actions'
import { ActionDetail } from './ActionDetail'
import { Roles } from './Roles'
import { Policies } from './Policies'
import { PolicyDetail } from './PolicyDetail'
import { Installs } from './Installs'
import { Readme } from './Readme'

export const appRoutes: RouteObject[] = [
  {
    element: <AppLayout />,
    children: [
      { path: ':orgId/apps/:appId', element: <Overview /> },
      { path: ':orgId/apps/:appId/components', element: <Components /> },
      { path: ':orgId/apps/:appId/components/:componentId', element: <ComponentDetail /> },
      {
        path: ':orgId/apps/:appId/components/:componentId/builds',
        loader: ({ params }) => redirect(`/${params.orgId}/apps/${params.appId}/components/${params.componentId}`),
      },
      { path: ':orgId/apps/:appId/components/:componentId/builds/:buildId', element: <BuildDetail /> },
      { path: ':orgId/apps/:appId/actions', element: <Actions /> },
      { path: ':orgId/apps/:appId/actions/:actionId', element: <ActionDetail /> },
      { path: ':orgId/apps/:appId/roles', element: <Roles /> },
      { path: ':orgId/apps/:appId/policies', element: <Policies /> },
      { path: ':orgId/apps/:appId/policies/:policyId', element: <PolicyDetail /> },
      { path: ':orgId/apps/:appId/installs', element: <Installs /> },
      { path: ':orgId/apps/:appId/readme', element: <Readme /> },
    ],
  },
]
