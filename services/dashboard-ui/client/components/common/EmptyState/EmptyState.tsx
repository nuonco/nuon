import React from 'react'
import type { TEmptyVariant } from '@/types'
import { cn } from '@/utils/classnames'
import { Text } from '../Text'
import { EmptyGraphic } from './EmptyGraphic'

export interface IEmptyState extends React.HTMLAttributes<HTMLDivElement> {
  action?: React.ReactNode
  emptyTitle?: string
  emptyMessage?: string
  isDarkModeOnly?: boolean
  size?: 'default' | 'sm'
  variant?: TEmptyVariant
}

export const EmptyState = ({
  action,
  className,
  emptyMessage = 'Nothing found',
  emptyTitle = 'Nothing to show',
  isDarkModeOnly = false,
  size = 'default',
  variant,
  ...props
}: IEmptyState) => {
  return (
    <div
      className={cn(
        'mx-auto my-6 flex flex-col items-center gap-2 w-full',
        {
          'max-w-56': size === 'default',
          'max-w-38': size === 'sm',
        },
        className
      )}
      {...props}
    >
      <EmptyGraphic
        variant={variant}
        isDarkModeOnly={isDarkModeOnly}
        size={size}
      />
      <span className="flex flex-col gap-1 items-center">
        <Text variant="subtext" weight="strong">
          {emptyTitle}
        </Text>
        <Text variant="label" className="text-center" theme="neutral">
          {emptyMessage}
        </Text>
      </span>

      {action ? <div className="mt-2">{action}</div> : null}
    </div>
  )
}
