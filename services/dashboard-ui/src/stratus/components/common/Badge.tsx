import React from 'react'
import { cn } from '@/stratus/components/helpers'
import './Badge.css'

type TBadgeVariant = 'default' | 'code'
type TBadgeTheme = 'neutral' | 'success' | 'warn' | 'error' | 'info'

interface IBadge extends React.HTMLAttributes<HTMLSpanElement> {
  size?: 'sm' | 'md' | 'lg'
  theme?: TBadgeTheme
  variant?: TBadgeVariant
}

export const Badge = ({
  className,
  children,
  size = 'lg',
  theme = 'neutral',
  variant = 'default',
  ...props
}: IBadge) => {
  return (
    <span className={cn('badge', variant, theme, size, className)} {...props}>
      {children}
    </span>
  )
}
