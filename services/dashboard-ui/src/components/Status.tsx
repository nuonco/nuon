'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { ToolTip } from '@/components/ToolTip'
import { Text } from '@/components/Typography'
import { titleCase } from '@/utils'

export type TStatus = 'active' | 'failed' | 'error' | 'waiting'

export interface IStatus {
  description?: string | false
  isStatusTextHidden?: boolean
  isLabelStatusText?: boolean
  label?: string | false
  status?: TStatus | string
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
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
          {label}
        </Text>
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
