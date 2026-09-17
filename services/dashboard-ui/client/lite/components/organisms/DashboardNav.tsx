import { cn } from '@/utils/classnames'
import { NavLink, type INavItem } from '../molecules/NavLink'

export interface IDashboardNav {
  items: INavItem[]
  label: string
  collapsed?: boolean
  onNavigate?: () => void
  className?: string
}

export const DashboardNav = ({
  items,
  label,
  collapsed = false,
  onNavigate,
  className,
}: IDashboardNav) => (
  <nav aria-label={label} className={cn('flex flex-col gap-1', className)}>
    {items.map((item) => (
      <NavLink
        key={`${item.label}-${item.href}`}
        {...item}
        collapsed={collapsed}
        onNavigate={onNavigate}
      />
    ))}
  </nav>
)
