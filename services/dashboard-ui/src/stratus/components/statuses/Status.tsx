import classNames from 'classnames'
import React, { type FC } from 'react'
import { removeKebabCase, sentanceCase } from '@/utils'
import { getStatusTheme, getStatusIcon } from './status-helpers'
import './Status.css'

export type TStatusType = string | 'success' | 'error'
type TStatusVariant = 'default' | 'badge' | 'timeline'

export interface IStatus
  extends Omit<React.HTMLAttributes<HTMLSpanElement>, 'children'> {
  isWithoutText?: boolean
  status: TStatusType
  variant?: TStatusVariant
}

export const Status: FC<IStatus> = ({
  className,
  isWithoutText = false,
  status,
  variant = 'default',
  ...props
}) => {
  const theme = getStatusTheme(status)
  const Icon = variant === 'timeline' ? getStatusIcon(status) : null

  return (
    <span
      className={classNames(`status ${variant}`, {
        [`${className}`]: Boolean(className),
      })}
      {...props}
    >
      <span className={`status-indicator ${theme}`}>
        {Icon ? (
          <Icon className={`status-icon`} weight="bold" size="18" />
        ) : null}
      </span>

      {isWithoutText ? null : (
        <span className="status-text">
          {sentanceCase(removeKebabCase(status))}
        </span>
      )}
    </span>
  )
}
