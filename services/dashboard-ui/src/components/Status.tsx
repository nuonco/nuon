'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { GoDotFill } from 'react-icons/go'
import { Text, ToolTip } from '@/components'
import { sentanceCase, titleCase } from '@/utils'

export type TStatus = 'active' | 'failed' | 'error' | 'waiting'

export interface IStatus {
  description?: string | false
  isStatusTextHidden?: boolean
  isLabelStatusText?: boolean
  label?: string | false
  status?: TStatus | string
}

export const Status: FC<IStatus> = ({
  description,
  isLabelStatusText = false,
  isStatusTextHidden = false,
  label,
  status = 'waiting',
}) => {
  const isActive = status === 'active' || status === 'ok'
  const isError =
    status === 'failed' ||
    status === 'error' ||
    status === 'bad' ||
    status === 'access-error' ||
    status === 'access_error'
  const isNoop = status === 'noop'

  return (
    <span className="flex flex-col gap-0">
      {label && !isLabelStatusText ? (
        <Text variant="label">{label}</Text>
      ) : null}
      <span
        className={classNames('flex gap-0 items-center', {
          'text-green-800 dark:text-green-500': isActive,
          'text-red-800 dark:text-red-500': isError,
          'text-cool-grey-600 dark:text-cool-grey-500': isNoop,
          'text-orange-800 dark:text-orange-500':
            !isActive && !isError && !isNoop,
        })}
      >
        <GoDotFill className="text-lg" />
        {isStatusTextHidden || isLabelStatusText ? null : (
          <Text variant="status">{status}</Text>
        )}
        {isLabelStatusText ? <Text variant="status">{label}</Text> : null}
      </span>
      {description ? (
        <Text className="truncate" variant="caption">
          {sentanceCase(description)}
        </Text>
      ) : null}
    </span>
  )
}

// TODO(nnnnat): rename and remove old status
export interface IStatusBadge extends IStatus {
  descriptionAlignment?: 'center' | 'left' | 'right'
  descriptionPosition?: 'bottom' | 'top'
  isWithoutBorder?: boolean
}

export const StatusBadge: FC<IStatusBadge> = ({
  description,
  descriptionAlignment,
  descriptionPosition,
  isLabelStatusText = false,
  isWithoutBorder = false,
  label,
  status,
}) => {
  const isActive = status === 'active' || status === 'ok'
  const isError =
    status === 'failed' ||
    status === 'error' ||
    status === 'bad' ||
    status === 'access-error' ||
    status === 'access_error'
  const isNoop = status === 'noop'
  const statusText = isLabelStatusText ? label : status

  const Status = (
    <span
      className={classNames('flex gap-2 items-center justify-start w-fit', {
        'border rounded-full pr-2 pl-1.5 py-0.5': !isWithoutBorder,
      })}
    >
      <span
        className={classNames('w-1.5 h-1.5 rounded-full', {
          'bg-green-800 dark:bg-green-500': isActive,
          'bg-red-800 dark:bg-red-500': isError,
          'bg-cool-grey-600 dark:bg-cool-grey-500': isNoop,
          'bg-orange-800 dark:bg-orange-500': !isActive && !isError && !isNoop,
        })}
      />
      <span className="text-sm font-medium">
        {titleCase(statusText as string)}
      </span>
    </span>
  )

  return (
    <span className="flex flex-col gap-2">
      {label && !isLabelStatusText ? (
        <span className="text-sm tracking-wide text-cool-grey-600 dark:text-cool-grey-500">
          {label}
        </span>
      ) : null}
      {description ? (
        <ToolTip
          alignment={descriptionAlignment}
          position={descriptionPosition}
          tipContent={description}
        >
          {Status}
        </ToolTip>
      ) : (
        Status
      )}
    </span>
  )
}
