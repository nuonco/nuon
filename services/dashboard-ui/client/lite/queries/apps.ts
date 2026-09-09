import { getApps } from '@/lib'
import { listQueryKey } from '../utils/list-query'

export const APPS_PAGE_SIZE = 20

export interface IAppsListQuery {
  orgId: string
  search?: string
  offset?: number
  pageSize?: number
}

export const appsListQuery = ({
  orgId,
  search = '',
  offset = 0,
  pageSize = APPS_PAGE_SIZE,
}: IAppsListQuery) => ({
  queryKey: [
    'apps',
    orgId,
    ...listQueryKey({ search, offset, pageSize, filters: {}, config: {} }),
  ],
  queryFn: () =>
    getApps({
      orgId,
      q: search || undefined,
      offset,
      limit: pageSize,
    }),
})
