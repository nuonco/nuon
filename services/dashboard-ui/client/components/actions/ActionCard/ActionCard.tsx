import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TActionConfigTriggerType } from '@/types'
import { ActionTriggerType } from '../ActionTriggerType'

export interface IActionCard {
  name?: string
  triggerType?: TActionConfigTriggerType
  status?: string
  href?: string
  isLoading?: boolean
  error?: string
  hasRun?: boolean
  canRun?: boolean
  onRun?: () => void
}

export const ActionCard = ({
  name,
  triggerType,
  status,
  href,
  isLoading,
  error,
  hasRun,
  canRun,
  onRun,
}: IActionCard) => {
  if (isLoading) {
    return (
      <div className="flex w-fit items-center gap-3 rounded-lg border px-3 py-2.5">
        <Skeleton width="6rem" />
        <Skeleton width="4rem" />
        <Skeleton width="3.5rem" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex w-fit items-center gap-2 rounded-lg border !border-red-200 dark:!border-red-900 px-3 py-2.5">
        <Text variant="subtext" theme="error">
          {error}
        </Text>
      </div>
    )
  }

  const content = (
    <div className="flex w-fit items-center gap-3 rounded-lg border px-3 py-2.5">
      {name && (
        canRun && href
          ? <Link href={href} className="text-sm">{name}</Link>
          : <Text variant="body" className="font-strong">{name}</Text>
      )}
      {hasRun && triggerType && <ActionTriggerType triggerType={triggerType} size="sm" />}
      {hasRun && status && <Status status={status} variant="badge" />}
      {!hasRun && (
        <Text variant="subtext" theme="neutral">
          No runs yet
        </Text>
      )}
      {canRun && onRun && (
        <Button size="sm" variant="primary" onClick={() => onRun()}>
          Run <Icon variant="PlayIcon" size={14} />
        </Button>
      )}
    </div>
  )

  if (href && !canRun) {
    return (
      <Link href={href} variant="ghost" className="flex !p-0 no-underline">
        {content}
      </Link>
    )
  }

  return content
}
