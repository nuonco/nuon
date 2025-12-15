import { api } from '@/lib/api'
import type { TOTELLog } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getLogStreamLogs = ({
  logStreamId,
  orgId,
  offset,
  order = 'asc',
}: {
  logStreamId: string
  orgId: string
  offset?: string
  order?: 'asc' | 'desc'
}) =>
  api<TOTELLog[]>({
    path: `log-streams/${logStreamId}/logs${buildQueryParams({ order })}`,
    orgId,
    headers: offset ? { 'X-Nuon-API-Offset': offset } : {},
  })
