import classNames from 'classnames'
import React, { type FC } from 'react'

export interface IBadge extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'default' | 'code' | 'status'
}

export const Badge: FC<IBadge> = ({
  className,
  children,
  variant = 'default',
  ...props
}) => {
  return (
    <span
      className={classNames(
        'text-sm bg-cool-grey-500/10 px-2 py-1 border flex items-center justify-start gap-2 w-fit h-fit leading-normal',
        {
          'rounded-full': variant === 'default',
          'rounded-lg font-mono tracking-wide': variant === 'code',
          'rounded-full !bg-transparent': variant === 'status',
          [`${className}`]: Boolean(className),
        }
      )}
      {...props}
    >
      {children}
    </span>
  )
}
