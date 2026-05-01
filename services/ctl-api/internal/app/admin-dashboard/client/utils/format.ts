import { DateTime, Duration } from 'luxon'

export function formatDate(dateStr: string | undefined): string {
  if (!dateStr) return '-'
  const dt = DateTime.fromISO(dateStr)
  return dt.isValid ? dt.toFormat('yyyy-MM-dd HH:mm:ss') : dateStr
}

export function formatRelativeDate(dateStr: string | undefined): string {
  if (!dateStr) return '-'
  const dt = DateTime.fromISO(dateStr)
  if (!dt.isValid) return dateStr
  return dt.toRelative() ?? dt.toFormat('yyyy-MM-dd HH:mm:ss')
}

export function formatDuration(nanos: number | undefined): string {
  if (!nanos) return '-'
  const dur = Duration.fromMillis(nanos / 1_000_000)
  if (dur.as('seconds') < 1) return `${Math.round(dur.as('milliseconds'))}ms`
  if (dur.as('minutes') < 1) return `${Math.round(dur.as('seconds'))}s`
  if (dur.as('hours') < 1) return dur.toFormat("m'm' s's'")
  return dur.toFormat("h'h' m'm'")
}

export function truncateId(id: string, len = 12): string {
  if (!id) return '-'
  return id.length > len ? id.slice(0, len) + '...' : id
}

export function statusColor(status: string | undefined): string {
  if (!status) return 'bg-gray-100 text-gray-700'
  const s = status.toLowerCase()
  if (s.includes('running') || s.includes('active') || s.includes('online') || s.includes('healthy') || s.includes('completed') || s.includes('success')) {
    return 'bg-green-100 text-green-800'
  }
  if (s.includes('failed') || s.includes('error') || s.includes('offline') || s.includes('unhealthy')) {
    return 'bg-red-100 text-red-800'
  }
  if (s.includes('pending') || s.includes('queued') || s.includes('waiting')) {
    return 'bg-yellow-100 text-yellow-800'
  }
  if (s.includes('cancel')) {
    return 'bg-orange-100 text-orange-800'
  }
  return 'bg-gray-100 text-gray-700'
}
