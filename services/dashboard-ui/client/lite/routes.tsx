import type { RouteObject } from 'react-router'
import { Home } from './views/Home'

export const liteRoutes: RouteObject[] = [
  { path: '/', element: <Home /> },
  { path: '*', element: <Home /> },
]
