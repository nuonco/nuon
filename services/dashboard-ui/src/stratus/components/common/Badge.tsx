import classNames from 'classnames'
import React, { type FC } from 'react'
import './Badge.css'

type TBadgeVariant = 'default' | 'code'
type TBadgeTheme = 'neutral' | 'success' | 'warn' | 'error' | 'info'

interface IBadge extends React.HTMLAttributes<HTMLSpanElement> {
  size?: 'sm' | 'md' | 'lg'
  theme?: TBadgeTheme
  variant?: TBadgeVariant
}

export const Badge: FC<IBadge> = ({
  className,
  children,
  size = 'lg',
  theme = 'neutral',
  variant = 'default',
  ...props
}) => {
  return (
    <span
      className={classNames(`badge ${variant} ${theme} ${size}`, {
        [`${className}`]: Boolean(className),
      })}
      {...props}
    >
      {children}
    </span>
  )
}
