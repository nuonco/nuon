import { Cron } from '@/components/common/Cron'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TActionConfigTriggerType } from '@/types'

const COMPONENT_ACTION_TRIGGERS: TActionConfigTriggerType[] = [
  'pre-deploy-component',
  'pre-teardown-component',
  'post-deploy-component',
  'post-teardown-component',
]

const ALL_COMPONENT_ACTION_TRIGGERS: TActionConfigTriggerType[] = [
  'pre-deploy-all-components',
  'pre-teardown-all-components',
  'post-deploy-all-components',
  'post-teardown-all-components',
]

const TRIGGER_LABELS: Partial<Record<TActionConfigTriggerType, string>> = {
  manual: 'Manual',
  cron: 'Cron',
  'pre-deploy-component': 'Pre-deploy',
  'pre-teardown-component': 'Pre-teardown',
  'pre-deploy-all-components': 'Pre-deploy',
  'pre-teardown-all-components': 'Pre-teardown',
  'post-deploy-component': 'Post-deploy',
  'post-teardown-component': 'Post-teardown',
  'post-deploy-all-components': 'Post-deploy',
  'post-teardown-all-components': 'Post-teardown',
}

export interface IActionTriggerType {
  componentName?: string
  componentPath?: string
  /**
   * Kept for backwards compatibility; no longer renders a badge.
   * @deprecated
   */
  size?: 'sm' | 'md' | 'lg'
  triggerType: TActionConfigTriggerType
  cronSchedule?: string
}

export const ActionTriggerType = ({
  componentName,
  componentPath,
  cronSchedule,
  triggerType,
}: IActionTriggerType) => {
  const label = TRIGGER_LABELS[triggerType] ?? triggerType
  const isComponentTrigger = COMPONENT_ACTION_TRIGGERS.includes(triggerType)
  const isAllComponents = ALL_COMPONENT_ACTION_TRIGGERS.includes(triggerType)
  const isCron = triggerType === 'cron'

  if (isCron) {
    return (
      <Text
        variant="subtext"
        theme="neutral"
        className="inline-flex items-center gap-1.5 min-w-0"
      >
        {label}:
        <Cron
          cron={cronSchedule}
          variant="subtext"
          family="mono"
          showTooltip
        />
      </Text>
    )
  }

  if (isAllComponents) {
    return (
      <Text variant="subtext" theme="neutral">
        {label}: <span className="font-mono">all components</span>
      </Text>
    )
  }

  if (isComponentTrigger && componentName) {
    return (
      <Text
        variant="subtext"
        theme="neutral"
        className="inline-flex items-center gap-1.5 min-w-0"
      >
        {label}:
        <Text variant="subtext" family="mono" className="truncate min-w-0">
          {componentPath ? (
            <Link href={componentPath}>{componentName}</Link>
          ) : (
            componentName
          )}
        </Text>
      </Text>
    )
  }

  return (
    <Text variant="subtext" theme="neutral">
      {label}
    </Text>
  )
}
