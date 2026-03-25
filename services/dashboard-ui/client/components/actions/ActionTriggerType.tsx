import { Badge, type IBadge } from '@/components/common/Badge'
import { Cron } from '@/components/common/Cron'
import { Link } from '@/components/common/Link'
import type { TActionConfigTriggerType } from '@/types'

const COMPONENT_ACTION_TRIGGERS: TActionConfigTriggerType[] = [
  'pre-deploy-component',
  'pre-teardown-component',
  'pre-deploy-all-components',
  'pre-teardown-all-components',
  'post-deploy-component',
  'post-teardown-component',
  'post-deploy-all-components',
  'post-teardown-all-components',
]

export interface IActionTriggerType {
  componentName?: string
  componentPath?: string
  size?: IBadge['size']
  triggerType: TActionConfigTriggerType
  cronSchedule?: string
}

export const ActionTriggerType = ({
  componentName,
  componentPath,
  cronSchedule,
  size,
  triggerType,
}: IActionTriggerType) => {
  const isComponentTrigger = COMPONENT_ACTION_TRIGGERS.includes(triggerType)
  const isCron = triggerType === 'cron'

  return (
    <Badge variant="code" size={size}>
      {isComponentTrigger ? (
        <span className="flex items-center gap-1">
          <span className="text-nowrap truncate">{triggerType}</span>:
          <Link href={componentPath}>{componentName}</Link>
        </span>
      ) : isCron ? (
        <span className="flex items-center gap-1">
          <span className="text-nowrap truncate">{triggerType}</span>:
          <Cron
            cron={cronSchedule}
            variant="label"
            theme="neutral"
            family="mono"
            showTooltip
          />
        </span>
      ) : (
        triggerType
      )}
    </Badge>
  )
}
