import { api } from '@/lib/api'
import type { TOTELLog } from '@/types'

export const getLogsByLogStreamId = ({
  logStreamId,
  orgId,
  offset,
}: {
  logStreamId: string
  orgId: string
  offset?: string
}) =>
  api<TOTELLog[]>({
    path: `log-streams/${logStreamId}/logs`,
    orgId,
    headers: offset ? { 'X-Nuon-API-Offset': offset } : {},
  })
