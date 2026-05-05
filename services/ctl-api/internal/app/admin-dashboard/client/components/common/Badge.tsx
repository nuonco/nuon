import cn from 'classnames'

interface IBadge {
  children: React.ReactNode
  variant?: 'default' | 'status'
  status?: string
  className?: string
}

function statusColor(status: string | undefined): string {
  if (!status) return 'bg-gray-100 text-gray-600 border-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:border-gray-700'
  const s = status.toLowerCase()
  if (s.includes('running') || s.includes('active') || s.includes('online') || s.includes('healthy') || s.includes('completed') || s.includes('success') || s === 'yes' || s === 'true') {
    return 'bg-green-50 text-green-700 border-green-200 dark:bg-green-900/30 dark:text-green-400 dark:border-green-800'
  }
  if (s.includes('failed') || s.includes('error') || s.includes('offline') || s.includes('unhealthy') || s === 'no' || s === 'false') {
    return 'bg-red-50 text-red-700 border-red-200 dark:bg-red-900/30 dark:text-red-400 dark:border-red-800'
  }
  if (s.includes('pending') || s.includes('queued') || s.includes('waiting')) {
    return 'bg-orange-50 text-orange-700 border-orange-200 dark:bg-orange-900/30 dark:text-orange-400 dark:border-orange-800'
  }
  if (s.includes('cancel')) {
    return 'bg-orange-50 text-orange-600 border-orange-200 dark:bg-orange-900/30 dark:text-orange-400 dark:border-orange-800'
  }
  return 'bg-gray-100 text-gray-600 border-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:border-gray-700'
}

export const Badge = ({ children, variant = 'default', status, className }: IBadge) => {
  const base = 'inline-flex items-center rounded-full border px-2 py-0.5 text-[11px] font-medium leading-4'
  const color = variant === 'status' && status
    ? statusColor(status)
    : 'bg-primary-50 text-primary-700 border-primary-200 dark:bg-primary-950 dark:text-primary-400 dark:border-primary-800'

  return <span className={cn(base, color, className)}>{children}</span>
}
