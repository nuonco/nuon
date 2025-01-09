import classNames from 'classnames'
import React, { type FC } from 'react'
import type { TActionConfigTriggerType } from '@/types'

interface IActionTriggerType extends React.HTMLAttributes<HTMLSpanElement> {
  triggerType: TActionConfigTriggerType | string
}

export const ActionTriggerType: FC<IActionTriggerType> = ({
  className,
  triggerType,
}) => {
  return (
    <span
      className={classNames(
        'px-2 py-1 border bg-gray-500/10 rounded-lg leading-none text-sm font-mono w-fit',
        {
          [`${className}`]: Boolean(className),
        }
      )}
    >
      {triggerType}
    </span>
  )
}
