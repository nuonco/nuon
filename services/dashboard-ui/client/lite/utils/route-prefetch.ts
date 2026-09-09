import type { QueryClient } from '@tanstack/react-query'
import { matchPath } from 'react-router'
import { appsListQuery } from '../queries/apps'
import { installsListQuery } from '../queries/installs'

type TRouteParams = Record<string, string | undefined>

interface IPrefetchableRoute {
  path: string
  queries: (
    params: TRouteParams
  ) => Array<{ queryKey: unknown[]; queryFn: () => Promise<unknown> }>
}

const PREFETCHABLE_ROUTES: IPrefetchableRoute[] = [
  {
    path: '/:orgId/apps',
    queries: ({ orgId }) => (orgId ? [appsListQuery({ orgId })] : []),
  },
  {
    path: '/:orgId/installs',
    queries: ({ orgId }) => (orgId ? [installsListQuery({ orgId })] : []),
  },
]

export const routePrefetchQueries = (href: string) => {
  if (!href.startsWith('/') || href.includes('?') || href.includes('#')) {
    return []
  }

  for (const route of PREFETCHABLE_ROUTES) {
    const match = matchPath(route.path, href)
    if (match) return route.queries(match.params)
  }

  return []
}

export const prefetchRoute = (queryClient: QueryClient, href: string) => {
  for (const query of routePrefetchQueries(href)) {
    void queryClient.prefetchQuery(query)
  }
}
