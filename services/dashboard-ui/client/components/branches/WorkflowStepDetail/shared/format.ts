export function statusTheme(status?: string) {
  if (status === 'success' || status === 'succeeded') return 'success'
  if (status === 'error') return 'error'
  if (status === 'in-progress') return 'info'
  return 'neutral'
}

export const formatDuration = (ns?: number | null): string => {
  if (!ns) return ''
  const secs = ns / 1_000_000_000
  if (secs < 60) return `${secs.toFixed(1)}s`
  const mins = Math.floor(secs / 60)
  const rem = Math.round(secs % 60)
  return `${mins}m ${rem}s`
}

export function getInitials(name?: string): string {
  if (!name) return '??'
  return name
    .split(' ')
    .map((w) => w[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)
}

export function cacheBadgeTheme(cacheStatus?: string) {
  if (cacheStatus === 'cache hit') return 'success'
  if (cacheStatus === 'no cache') return 'warn'
  return 'neutral'
}

export function diffRowBg(op?: string) {
  if (op === 'add') return 'bg-green-500/[0.06]'
  if (op === 'remove') return 'bg-red-500/[0.06]'
  return ''
}
