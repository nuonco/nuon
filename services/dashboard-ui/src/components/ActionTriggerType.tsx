import classNames from 'classnames'
import React, { type FC } from 'react'
import { Badge } from '@/components/Badge'
import { Link } from '@/components/Link'
import type { TActionConfigTriggerType } from '@/types'

interface IActionTriggerType extends React.HTMLAttributes<HTMLSpanElement> {
  componentName?: string
  componentPath?: string
  triggerType: TActionConfigTriggerType | string
}

export const ActionTriggerType: FC<IActionTriggerType> = ({
  className,
  componentName,
  componentPath,
  triggerType,
}) => {
  return (
    <Badge
      className={classNames('inline-flex gap-1', {
        [`${className}`]: Boolean(className),
      })}
      variant="code"
    >
      {triggerType}
      {(triggerType === 'pre-deploy-component' ||
        triggerType === 'post-deploy-component') &&
      componentName ? (
        componentPath ? (
          <>
            :
            <Link className="inline-flex" href={componentPath}>
              {componentName}
            </Link>
          </>
        ) : (
          `: ${componentName}`
        )
      ) : null}
    </Badge>
  )
}
