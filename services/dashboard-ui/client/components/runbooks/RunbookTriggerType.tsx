import { Cron } from '@/components/common/Cron'
import { Text } from '@/components/common/Text'
import type { TRunbookTriggerType } from '@/lib/ctl-api/apps/runbooks'

interface IRunbookTriggerType {
  triggerType: TRunbookTriggerType
  cronSchedule?: string
}

export const RunbookTriggerType = ({
  cronSchedule,
  triggerType,
}: IRunbookTriggerType) => {
  if (triggerType === 'cron') {
    return (
      <Text
        variant="subtext"
        theme="neutral"
        className="inline-flex items-center gap-1.5 min-w-0"
      >
        Cron:
        <Cron cron={cronSchedule} variant="subtext" family="mono" showTooltip />
      </Text>
    )
  }

  return (
    <Text variant="subtext" theme="neutral">
      Manual
    </Text>
  )
}
