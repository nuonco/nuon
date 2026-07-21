import { api } from '@/lib/api'
import type { TBuild } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export type TOrgComponentBuildHistoryItem = {
  build: TBuild
  app_id: string
  component_id: string
  component_name: string
  build_runner_job_id: string | null
}

export type TOrgComponentBuildHistoryResponse = {
  items: TOrgComponentBuildHistoryItem[]
  next_cursor: string | null
  previous_cursor: string | null
}

export const getOrgComponentBuildHistory = ({
  cursor,
  limit = 10,
  orgId,
}: {
  cursor?: string
  limit?: number
  orgId: string
}) =>
  api<TOrgComponentBuildHistoryResponse>({
    path: `component-builds${buildQueryParams({ cursor, limit })}`,
    orgId,
  })
