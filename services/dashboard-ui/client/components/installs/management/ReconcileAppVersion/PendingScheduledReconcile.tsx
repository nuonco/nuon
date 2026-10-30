import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TInstallScheduledReconcile } from '@/types'

export interface IPendingScheduledReconcile {
  pending: TInstallScheduledReconcile
  isCanceling?: boolean
  onCancel: () => void
  size?: 'sm' | 'md'
}

export const PendingScheduledReconcile = ({
  pending,
  isCanceling,
  onCancel,
  size = 'md',
}: IPendingScheduledReconcile) => (
  <div className="flex flex-wrap items-center justify-between gap-2">
    <div className="flex items-center gap-2 min-w-0">
      <Icon variant="ClockIcon" size="16" />
      <Text variant={size === 'sm' ? 'subtext' : 'body'} className="truncate">
        Scheduled for{' '}
        <Time time={pending.scheduled_at} format="short-datetime" /> (
        <Time time={pending.scheduled_at} format="relative" />)
      </Text>
      {pending.plan_only ? (
        <Badge size="sm" theme="neutral">
          Plan only
        </Badge>
      ) : null}
    </div>
    <Button
      size="sm"
      variant="secondary"
      onClick={onCancel}
      disabled={isCanceling}
    >
      {isCanceling ? 'Canceling...' : 'Cancel schedule'}
    </Button>
  </div>
)
