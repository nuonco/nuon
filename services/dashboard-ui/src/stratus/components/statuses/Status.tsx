import React, { type FC } from 'react'
import { cn } from '@/stratus/components/helpers'
import { Icon } from '@/stratus/components/common'
import { removeKebabCase, sentanceCase } from '@/utils'
import { getStatusTheme, getStatusIconVariant } from './status-helpers'
import './Status.css'

export type TStatusType = string | 'success' | 'error'
type TStatusVariant = 'default' | 'badge' | 'timeline'

export interface IStatus
  extends Omit<React.HTMLAttributes<HTMLSpanElement>, 'children'> {
  children?: React.ReactNode
  isWithoutText?: boolean
  status: TStatusType
  variant?: TStatusVariant
}

export const Status: FC<IStatus> = ({
  children,
  className,
  isWithoutText = false,
  status,
  variant = 'default',
  ...props
}) => {
  const theme = getStatusTheme(status)
  const iconVariant =
    variant === 'timeline' ? getStatusIconVariant(status) : null

  return (
    <span className={cn('status', variant, className)} {...props}>
      <span className={`status-indicator ${theme}`}>
        {iconVariant ? (
          <Icon
            className="status-icon"
            variant={iconVariant}
            weight="bold"
            size="18"
          />
        ) : null}
      </span>

      {isWithoutText ? null : (
        <span className="status-text">
          {children || sentanceCase(removeKebabCase(status))}
        </span>
      )}
    </span>
  )
}
