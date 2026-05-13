import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Duration } from '@/components/common/Duration'
import { Time } from '@/components/common/Time'
import { LabeledValue } from '@/components/common/LabeledValue'
import type { TRunnerHeartbeat } from '@/types'

interface ILoadRunnerHeartbeat {
  heartbeat?: TRunnerHeartbeat
  error: string | null
  isLoading: boolean
  platform?: string
  isPlatformLoading: boolean
}

export const LoadRunnerHeartbeat = ({
  heartbeat,
  error,
  isLoading,
  platform,
  isPlatformLoading,
}: ILoadRunnerHeartbeat) => {
  if (error) {
    return (
      <Text variant="subtext" className="text-red-600">
        {error}
      </Text>
    )
  }

  if (isLoading) {
    return (
      <div className="flex items-center gap-2">
        <Icon variant="Loading" className="animate-spin" size="16" />
        <Text variant="subtext">Loading heartbeat...</Text>
      </div>
    )
  }

  if (!heartbeat?.version) {
    return <Text variant="subtext">No heartbeat yet</Text>
  }

  return (
    <div className="flex items-start gap-4">
      <LabeledValue label="Version">{heartbeat.version}</LabeledValue>
      <LabeledValue label="Alive time">
        <div className="flex items-center gap-1">
          <Icon variant="TimerIcon" />
          <Duration variant="subtext" nanoseconds={heartbeat.alive_time} />
        </div>
      </LabeledValue>
      <LabeledValue label="Last heartbeat seen">
        <span className="flex items-center gap-1">
          <Icon variant="HeartbeatIcon" />
          <Time
            variant="subtext"
            time={heartbeat.created_at}
            format="relative"
          />
        </span>
      </LabeledValue>

      <LabeledValue label="Platform">
        {isPlatformLoading ? (
          <div className="flex items-center gap-2">
            <Icon variant="Loading" className="animate-spin" size="16" />
            <Text variant="subtext">Loading platform...</Text>
          </div>
        ) : (
          platform
        )}
      </LabeledValue>
    </div>
  )
}
