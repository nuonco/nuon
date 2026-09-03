import type { HTMLAttributes, ReactNode } from 'react'
import { cn } from '@/utils/classnames'
import { Card } from '../atoms/Card'
import { Link } from '../atoms/Link'
import { Logo } from '../atoms/Logo'

export interface IFocusHeader
  extends Omit<HTMLAttributes<HTMLElement>, 'children'> {
  actions?: ReactNode
  homeHref?: string
  scrolled?: boolean
}

export const FocusHeader = ({
  actions,
  homeHref = '/',
  scrolled = false,
  className,
  ...props
}: IFocusHeader) => (
  <Card
    as="header"
    padding="none"
    blur={scrolled ? 'lg' : 'none'}
    shadow={scrolled ? 'floating' : 'none'}
    className={cn(
      'sticky top-3 z-30 flex h-16 shrink-0 items-center justify-between gap-4 px-4 transition-[background-color,border-color,box-shadow,backdrop-filter] sm:px-6 lg:px-8',
      !scrolled && 'border-transparent !bg-transparent backdrop-blur-none',
      className
    )}
    {...props}
  >
    <Link
      href={homeHref}
      aria-label="Nuon home"
      className="block shrink-0"
      style={{ color: 'var(--text-primary)' }}
    >
      <Logo variant="wordmark" tone="mono" size={24} />
    </Link>
    {actions ? (
      <div className="flex min-w-0 items-center justify-end gap-2">
        {actions}
      </div>
    ) : null}
  </Card>
)
