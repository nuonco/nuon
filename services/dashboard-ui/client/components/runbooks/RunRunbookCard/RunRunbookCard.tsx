import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'

export interface IRunRunbookCard {
  name?: string
  href?: string
  stepCount?: number
  isLoading?: boolean
  error?: string
  onRun?: () => void
}

export const RunRunbookCard = ({
  name,
  href,
  stepCount,
  isLoading,
  error,
  onRun,
}: IRunRunbookCard) => {
  if (isLoading) {
    return (
      <div className="flex w-fit items-center gap-3 rounded-lg border px-3 py-2.5">
        <Skeleton width="8rem" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex w-fit items-center gap-3 rounded-lg border border-red-300 dark:border-red-700 px-3 py-2.5">
        <Text variant="subtext" className="text-red-600 dark:text-red-400">
          {error}
        </Text>
      </div>
    )
  }

  return (
    <div className="flex w-fit items-center gap-3 rounded-lg border px-3 py-2.5">
      <Icon variant="BookIcon" size={16} className="text-cool-grey-500" />
      {href ? <Link href={href} className="text-sm">{name}</Link> : <Text variant="body" className="text-sm">{name}</Text>}
      {typeof stepCount === 'number' && (
        <Badge size="sm" theme="neutral">
          {stepCount} step{stepCount !== 1 ? 's' : ''}
        </Badge>
      )}
      <Button size="sm" variant="primary" onClick={onRun}>
        Run <Icon variant="PlayIcon" size={14} />
      </Button>
    </div>
  )
}
