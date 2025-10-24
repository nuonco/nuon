// @ts-nocheck
import { DateTime } from 'luxon'
import type { TLogRecord } from './types'
import type { TOTELLog } from '@/types'

function isoToEpochNanos(iso: string): bigint {
  const dt = DateTime.fromISO(iso)
  const msSinceEpoch = BigInt(dt.toMillis())

  // Manually extract nanoseconds
  const match = iso.match(/\.(\d+)Z$/)
  let nanos = 0n
  if (match) {
    // Pad to 9 digits for nanosecond precision
    nanos = BigInt((match[1] + '000000000').slice(0, 9))
  }

  // Combine ms part with nanoseconds
  return msSinceEpoch * 1000000n + nanos
}

export function parseOTELLog(logs: Array<TOTELLog>): Array<TLogRecord> {
  return logs?.length
    ? logs?.map((l) => ({
        ...l,
        timestamp: l.timestamp,
      }))
    : []
}
