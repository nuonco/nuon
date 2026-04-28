import { DateTime } from 'luxon'

export function isRecentTimestamp(
  timestampStr: string | undefined,
  maxAgeSeconds = 60,
) {
  if (!timestampStr) return false
  const date = DateTime.fromISO(timestampStr)
  const now = DateTime.now()
  const diffInSeconds = now.diff(date, 'seconds').seconds

  return diffInSeconds >= 0 && diffInSeconds < maxAgeSeconds
}
