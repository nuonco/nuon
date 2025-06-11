import classNames from 'classnames'
import React, { type FC } from 'react'
import { Tooltip, Text, type ITooltip } from '@/stratus/components/common'
import { Status, type IStatus } from './Status'

interface IDetailedStatus
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'children'> {
  status: IStatus
  tooltip: Omit<ITooltip, 'children'>
  title: string
}

export const DetailedStatus: FC<IDetailedStatus> = ({
  className,
  status: { variant = 'badge', ...status },
  tooltip,
  title,
  ...props
}) => {
  return (
    <div
      className={classNames('flex flex-col gap-1', {
        [`${className}`]: Boolean(className),
      })}
      {...props}
    >
      <Text variant="subtext" theme="muted">
        {title}
      </Text>
      <Tooltip {...tooltip}>
        <Status variant={variant} {...status} />
      </Tooltip>
    </div>
  )
}
