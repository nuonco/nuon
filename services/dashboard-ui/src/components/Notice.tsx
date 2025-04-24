import classNames from 'classnames'
import React, { type FC } from 'react'
import { Warning, WarningOctagon } from '@phosphor-icons/react/dist/ssr'

interface INotice {
  className?: string
  children: React.ReactNode
  variant?: 'error' | 'warn' | 'info' | 'success'
}

export const Notice: FC<INotice> = ({
  className,
  children,
  variant = 'error',
}) => {
  const Icon =
    variant === 'warn' ? <Warning size="20" /> : <WarningOctagon size="18" />

  return (
    <div
      className={classNames(
        'flex items-center gap-2 justify-start w-full p-2 border rounded-md',
        {
          'border-red-400 bg-red-300/20 text-red-800 dark:border-red-600 dark:bg-red-600/5 dark:text-red-600':
            variant === 'error',
          'border-orange-400 bg-orange-300/20 text-orange-800 dark:border-orange-600 dark:bg-orange-600/5 dark:text-orange-600':
            variant === 'warn',
          [`${className}`]: Boolean(className),
        }
      )}
    >
      <span>{Icon}</span>
      <span className="leading-normal text-sm font-mono">{children}</span>
    </div>
  )
}
