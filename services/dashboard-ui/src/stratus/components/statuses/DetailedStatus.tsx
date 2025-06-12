import React, { type FC } from 'react'
import { LabeledValue, type ILabeledValue, Tooltip, type ITooltip } from '@/stratus/components/common'
import { Status, type IStatus } from './Status'

interface IDetailedStatus
  extends Omit<ILabeledValue, 'children'> {
  status: IStatus
  tooltip: Omit<ITooltip, 'children'>
}

export const DetailedStatus: FC<IDetailedStatus> = ({
  status: { variant = 'badge', ...status },
  tooltip: { position = 'bottom', ...tooltip},
  ...props
}) => {
  return (
    <LabeledValue {...props}>
      <Tooltip position={position} {...tooltip}>
        <Status variant={variant} {...status} />
      </Tooltip>
    </LabeledValue>
  )
}
