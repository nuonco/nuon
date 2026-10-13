import { useEffect, useState, type ReactNode } from 'react'
import {
  createMemoryRouter,
  RouterProvider,
  type RouteObject,
} from 'react-router'
import { useQueryClient, type QueryKey } from '@tanstack/react-query'
import { queryClient } from '@/lib/query-client'
import { orgRoutes } from '@/views/org/routes'
import {
  beginInstallFixture,
  endInstallFixture,
  type TFixture,
} from './install-fixture'

export const InstallView = ({
  fixture,
  path,
  initialQueryData = [],
  routeElements = {},
}: {
  fixture: TFixture
  path: string
  initialQueryData?: Array<{ queryKey: QueryKey; data: unknown }>
  routeElements?: Record<string, ReactNode>
}) => {
  const client = useQueryClient()
  const [session] = useState(() => {
    client.clear()
    queryClient.clear()
    for (const { queryKey, data } of initialQueryData) {
      client.setQueryData(queryKey, data)
      if (client !== queryClient) queryClient.setQueryData(queryKey, data)
    }
    const generation = beginInstallFixture(fixture)
    const previewRoutes = (routes: RouteObject[]): RouteObject[] =>
      routes.map(
        (route) =>
          ({
            ...route,
            element: routeElements[route.path || ''] ?? route.element,
            ...(route.children
              ? { children: previewRoutes(route.children) }
              : {}),
          }) as RouteObject
      )
    return {
      generation,
      router: createMemoryRouter(previewRoutes(orgRoutes), {
        initialEntries: [path],
      }),
    }
  })

  useEffect(
    () => () => {
      session.router.dispose()
      endInstallFixture(session.generation)
    },
    [session]
  )

  return <RouterProvider router={session.router} />
}
