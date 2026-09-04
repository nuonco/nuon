import type { HTMLAttributes } from 'react'
import { useLocation } from 'react-router'
import { cn } from '@/utils/classnames'
import { Link } from '../atoms/Link'

export interface ISubNavItem {
  href: string
  label: string
  end?: boolean
}

export interface ISubNav extends Omit<HTMLAttributes<HTMLElement>, 'children'> {
  items: ISubNavItem[]
  label: string
}

const LINK_CLASSES =
  'flex h-8 shrink-0 items-center rounded-lg px-3 text-caption text-secondary no-underline outline-none transition-colors ' +
  'hover:bg-surface-02 hover:text-primary active:bg-surface-03 ' +
  'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus-ring ' +
  'aria-[current=page]:bg-surface-02 aria-[current=page]:font-medium aria-[current=page]:text-primary'

export const SubNav = ({ items, label, className, ...props }: ISubNav) => {
  const { pathname } = useLocation()

  return (
    <nav
      aria-label={label}
      className={cn(
        'flex max-w-full items-center gap-1 overflow-x-auto',
        className
      )}
      {...props}
    >
      {items.map((item) => {
        const active = item.end
          ? pathname === item.href
          : pathname === item.href || pathname.startsWith(`${item.href}/`)

        return (
          <Link
            key={item.href}
            href={item.href}
            aria-current={active ? 'page' : undefined}
            className={LINK_CLASSES}
          >
            {item.label}
          </Link>
        )
      })}
    </nav>
  )
}
