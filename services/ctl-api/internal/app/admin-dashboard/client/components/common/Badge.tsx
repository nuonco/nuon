import cn from 'classnames'
import { statusColor } from '@/utils/format'

interface IBadge {
  children: React.ReactNode
  variant?: 'default' | 'status'
  status?: string
  className?: string
}

export const Badge = ({ children, variant = 'default', status, className }: IBadge) => {
  const base = 'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium'
  const color = variant === 'status' && status ? statusColor(status) : 'bg-gray-100 text-gray-700'

  return <span className={cn(base, color, className)}>{children}</span>
}
