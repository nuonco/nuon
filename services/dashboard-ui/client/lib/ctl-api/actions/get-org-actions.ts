import { api } from '@/lib/api'
import type { TAction, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export interface IGetOrgActions extends TPaginationParams {
  orgId: string
  q?: string
}

export const getOrgActions = ({ orgId, q, limit, offset }: IGetOrgActions) =>
  api<TAction[]>({
    orgId,
    path: `action-workflows${buildQueryParams({ limit, offset, q })}`,
    paginated: true,
  })
