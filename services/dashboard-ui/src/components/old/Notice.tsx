import classNames from 'classnames'
import React, { type FC } from 'react'
import {
  CheckCircle,
  Info,
  Warning,
  WarningOctagon,
} from '@phosphor-icons/react/dist/ssr'

export interface INotice {
  className?: string
  children: React.ReactNode
  variant?: 'error' | 'warn' | 'info' | 'success' | 'default'
}

export const Notice: FC<INotice> = ({
  className,
  children,
  variant = 'error',
}) => {
  const Icon =
    variant === 'warn' ? (
      <Warning size="20" />
    ) : variant === 'error' ? (
      <WarningOctagon size="20" />
    ) : variant === 'success' ? (
      <CheckCircle size="20" />
    ) : (
      <Info size="20" />
    )

  return (
    <div
      className={classNames('flex gap-4 w-full px-2 py-1 border rounded-md', {
        'bg-cool-grey-50 text-cool-grey-800 border-cool-grey-300 dark:bg-cool-grey-600/15 dark:border-cool-grey-600/40 dark:text-cool-grey-500':
          variant === 'default',
        'bg-blue-50 text-blue-800 border-blue-300 dark:bg-blue-600/15 dark:border-blue-600/40 dark:text-blue-500':
          variant === 'info',
        'bg-orange-50 text-orange-800 border-orange-300 dark:bg-orange-600/15 dark:border-orange-600/40 dark:text-orange-500':
          variant === 'warn',
        'bg-red-50 text-red-800 border-red-300 dark:bg-red-600/15 dark:border-red-600/40 dark:text-red-500':
          variant === 'error',
        'bg-green-50 text-green-800 border-green-300 dark:bg-green-600/15 dark:border-green-600/40 dark:text-green-500':
          variant === 'success',
        [`${className}`]: Boolean(className),
      })}
    >
      <span className="flex self-start">{Icon}</span>
      <span className="leading-normal text-sm font-mono py-0.5 !w-full">
        {children}
      </span>
    </div>
  )
}
