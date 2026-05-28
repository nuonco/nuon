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
  app_id?: string
  org_id?: string
  created_at?: string
  updated_at?: string
  labels?: Record<string, string>
  configs?: TRunbookConfig[]
}

export type TRunbookConfig = {
  id: string
  runbook_id?: string
  app_config_id?: string
  readme?: string
  steps?: TRunbookStep[]
  created_at?: string
}

export type TRunbookStep = {
  id: string
  name: string
  idx?: number
  type?: string
  component_name?: string
  deploy_dependencies?: boolean
  command?: string
  inline_contents?: string
  role?: string
  env_vars?: Record<string, string>
}
