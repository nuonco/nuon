import classNames from 'classnames'
import React, { type FC } from 'react'
import { GoDotFill } from 'react-icons/go'
import { Text }  from '@/components';
import { sentanceCase } from "@/utils"

export type TStatus = 'active' | 'failed' | 'error' | 'waiting'

export interface IStatus {
  description?: string | false;
  isStatusTextHidden?: boolean;
  isLabelStatusText?:boolean;
  label?: string | false;
  status?: TStatus | string;
}

export const Status: FC<IStatus> = ({ description, isLabelStatusText = false, isStatusTextHidden = false, label, status = 'waiting' }) => {
  const isActive = status === 'active' || status === "ok"
  const isError = status === 'failed' || status === 'error' || status === "bad"

  return (
    <span className="flex flex-col gap-0">
      {label && !isLabelStatusText ? (<Text variant="label">{label}</Text>) : null}
      <span
        className={classNames('flex gap-0 items-center', {
          'text-green-700 dark:text-green-500': isActive,
          'text-red-600 dark:text-red-500': isError,
          'text-yellow-600 dark:text-yellow-500': !isActive && !isError,
        })}
      >
        <GoDotFill className="text-lg" />
        {isStatusTextHidden || isLabelStatusText ? null : <Text variant="status">{status}</Text>}
        {isLabelStatusText ? <Text variant="status">{label}</Text> : null}
      </span>
      {description ? (<Text className="truncate" variant="caption">{sentanceCase(description)}</Text>) : null}
    </span>
  )
}
