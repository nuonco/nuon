import { api } from '@/lib/api'
import type { TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export interface IGetRunbooks extends TPaginationParams {
  appId: string
  orgId: string
}

export async function getRunbooks({ appId, orgId, limit, offset }: IGetRunbooks) {
  return api<TRunbook[]>({
    orgId,
    path: `apps/${appId}/runbooks${buildQueryParams({ limit, offset })}`,
    paginated: true,
  })
}

export type TRunbook = {
  id: string
  name: string
  description?: string
  readme?: string
  app_id?: string
  org_id?: string
  created_at?: string
  updated_at?: string
  steps?: TRunbookStep[]
}

export type TRunbookStep = {
  id: string
  name: string
  description?: string
  idx?: number
  type?: string
}
