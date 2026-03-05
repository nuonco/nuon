import { DateTime } from 'luxon'

// take ISO timestamp string and return true if it's less than 15 seconds ago
export function isLessThan15SecondsOld(timestampStr: string) {
  const date = DateTime.fromISO(timestampStr)
  const now = DateTime.now()
  const diffInSeconds = now.diff(date, 'seconds').seconds

  return diffInSeconds >= 0 && diffInSeconds < 15
}
