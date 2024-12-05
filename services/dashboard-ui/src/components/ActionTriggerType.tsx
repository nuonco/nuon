import React, { type FC } from 'react'
import type { TActionConfigTriggerType } from '@/types'

interface IActionTriggerType {
  triggerType: TActionConfigTriggerType
}

export const ActionTriggerType: FC<IActionTriggerType> = ({ triggerType }) => {
  return (
    <span className="p-2 border bg-gray-500/10 rounded-lg leading-none text-sm font-mono">
      {triggerType}
    </span>
  )
}
