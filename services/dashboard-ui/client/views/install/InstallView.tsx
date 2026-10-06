import { useEffect, useState } from 'react'
import { createMemoryRouter, RouterProvider } from 'react-router'
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
}: {
  fixture: TFixture
  path: string
  initialQueryData?: Array<{ queryKey: QueryKey; data: unknown }>
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
    return {
      generation,
      router: createMemoryRouter(orgRoutes, { initialEntries: [path] }),
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
