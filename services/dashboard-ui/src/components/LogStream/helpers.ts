import { DateTime } from 'luxon'
import type { TLogRecord } from './types'
import type { TOTELLog } from '@/types'

// convert otel log timestamp from string to milliseconds
export function parseOTELLog(logs: Array<TOTELLog>): Array<TLogRecord> {
  return logs?.length
    ? logs?.map((l) => ({
        ...l,
        timestamp: DateTime.fromISO(l.timestamp).toMillis(),
      }))
    : []
}
