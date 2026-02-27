import type { RouteObject } from 'react-router'
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

export const installRoutes: RouteObject[] = [
  {
    element: <InstallLayout />,
    children: [
      { path: ':orgId/installs/:installId', element: <Overview /> },
      { path: ':orgId/installs/:installId/components', element: <Components /> },
      { path: ':orgId/installs/:installId/actions', element: <Actions /> },
      { path: ':orgId/installs/:installId/roles', element: <Roles /> },
      { path: ':orgId/installs/:installId/policies', element: <Policies /> },
      { path: ':orgId/installs/:installId/runner', element: <Runner /> },
      { path: ':orgId/installs/:installId/sandbox', element: <Sandbox /> },
      { path: ':orgId/installs/:installId/stacks', element: <Stacks /> },
      { path: ':orgId/installs/:installId/workflows', element: <Workflows /> },
      { path: ':orgId/installs/:installId/readme', element: <Readme /> },
    ],
  },
]
