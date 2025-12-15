import { api } from '@/lib/api'
import type { TLogStream } from '@/types'

export const getLogStream = ({
  logStreamId,
  orgId,
}: {
  logStreamId: string
  orgId: string
}) =>
  api<TLogStream>({
    path: `log-streams/${logStreamId}`,
    orgId,
  })
