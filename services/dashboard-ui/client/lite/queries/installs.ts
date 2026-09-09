import { getInstalls } from '@/lib'
import {
  commaSetQueryParameter,
  listQueryKey,
  type TListQueryValues,
} from '../utils/list-query'

export const INSTALLS_PAGE_SIZE = 20

export const INSTALL_FILTERS = {
  labels: commaSetQueryParameter('labels'),
  branches: commaSetQueryParameter('branches'),
}

export type TInstallFilters = TListQueryValues<typeof INSTALL_FILTERS>

export const emptyInstallFilters = (): TInstallFilters => ({
  labels: new Set<string>(),
  branches: new Set<string>(),
})

export interface IInstallsListQuery {
  orgId: string
  search?: string
  offset?: number
  pageSize?: number
  filters?: TInstallFilters
}

export const installsListQuery = ({
  orgId,
  search = '',
  offset = 0,
  pageSize = INSTALLS_PAGE_SIZE,
  filters = emptyInstallFilters(),
}: IInstallsListQuery) => ({
  queryKey: [
    'installs',
    orgId,
    ...listQueryKey({
      search,
      offset,
      pageSize,
      filters,
      config: INSTALL_FILTERS,
    }),
  ],
  queryFn: () =>
    getInstalls({
      orgId,
      q: search || undefined,
      offset,
      limit: pageSize,
      labels: [...filters.labels].join(',') || undefined,
      branches: [...filters.branches].join(',') || undefined,
    }),
})
