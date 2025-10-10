import classNames from 'classnames'
import React, { type FC } from 'react'

export interface IBadge extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'default' | 'code' | 'status'
  theme?: 'default' | 'info' | 'warn' | 'error' | 'success'
  isCompact?: boolean
}

export const Badge: FC<IBadge> = ({
  className,
  children,
  isCompact = false,
  theme = 'default',
  variant = 'default',
  ...props
}) => {
  return (
    <span
      className={classNames(
        'text-xs px-2 py-1 border flex items-center justify-start gap-2 w-fit h-fit leading-normal',
        {
          'rounded-full': variant === 'default',
          'rounded-lg font-mono tracking-wide': variant === 'code',
          'rounded-full !bg-transparent': variant === 'status',
          'bg-cool-grey-50 text-cool-grey-800 border-cool-grey-300 dark:bg-cool-grey-600/15 dark:border-cool-grey-600/40 dark:text-cool-grey-500':
            theme === 'default',
          'bg-blue-50 text-blue-800 border-blue-300 dark:bg-blue-600/15 dark:border-blue-600/40 dark:text-blue-500':
            theme === 'info',
          'bg-orange-50 text-orange-800 border-orange-300 dark:bg-orange-600/15 dark:border-orange-600/40 dark:text-orange-500':
            theme === 'warn',
          'bg-red-50 text-red-800 border-red-300 dark:bg-red-600/15 dark:border-red-600/40 dark:text-red-500':
            theme === 'error',
          'bg-green-50 text-green-800 border-green-300 dark:bg-green-600/15 dark:border-green-600/40 dark:text-green-500':
            theme === 'success',
          '!py-0.5': isCompact,
          [`${className}`]: Boolean(className),
        }
      )}
      {...props}
    >
      {children}
    </span>
  )
}
