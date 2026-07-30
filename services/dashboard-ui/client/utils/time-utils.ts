import { DateTime } from 'luxon'

// Mirrors componentHealthStaleAfter in the health evaluator: past this, an
// observation no longer counts toward a component's verdict, so the UI must not
// present it as current.
export const HEALTH_OBSERVATION_STALE_AFTER_SECONDS = 5 * 60

export function isStaleObservation(
  timestampStr: string | undefined,
  maxAgeSeconds = HEALTH_OBSERVATION_STALE_AFTER_SECONDS,
) {
  if (!timestampStr) return false
  const seconds = DateTime.now().diff(DateTime.fromISO(timestampStr), 'seconds').seconds
  return seconds > maxAgeSeconds
}

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
