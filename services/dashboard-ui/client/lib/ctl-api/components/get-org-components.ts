import { api } from '@/lib/api'
import type { TComponent, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export interface IGetOrgComponents extends TPaginationParams {
  orgId: string
  q?: string
  component_ids?: string
}

export const getOrgComponents = ({
  orgId,
  q,
  component_ids,
  limit,
  offset,
}: IGetOrgComponents) =>
  api<TComponent[]>({
    orgId,
    path: `components${buildQueryParams({ limit, offset, q, component_ids })}`,
    paginated: true,
  })
