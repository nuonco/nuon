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
}

export const ActionCard = ({
  name,
  triggerType,
  status,
  href,
  isLoading,
  error,
  hasRun,
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
        <Text variant="body" className="font-strong">
          {name}
        </Text>
      )}
      {hasRun && triggerType && <ActionTriggerType triggerType={triggerType} size="sm" />}
      {hasRun && status && <Status status={status} variant="badge" />}
      {!hasRun && (
        <Text variant="subtext" theme="neutral">
          No runs yet
        </Text>
      )}
    </div>
  )

  if (href) {
    return (
      <Link href={href} variant="ghost" className="flex !p-0 no-underline">
        {content}
      </Link>
    )
  }

  return content
}
