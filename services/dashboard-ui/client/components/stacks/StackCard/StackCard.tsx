import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'

export interface IStackCard {
  status?: string
  runCount?: number
  createdAt?: string
  href?: string
  isLoading?: boolean
  error?: string
}

const Skeleton = () => (
  <div className="flex w-fit items-center gap-3 rounded-lg border border-cool-grey-200 dark:border-cool-grey-800 px-3 py-2.5 animate-pulse">
    <div className="h-4 w-12 rounded bg-cool-grey-200 dark:bg-cool-grey-800" />
    <div className="h-4 w-14 rounded bg-cool-grey-200 dark:bg-cool-grey-800" />
    <div className="h-4 w-12 rounded bg-cool-grey-200 dark:bg-cool-grey-800" />
    <div className="h-4 w-16 rounded bg-cool-grey-200 dark:bg-cool-grey-800" />
  </div>
)

export const StackCard = ({
  status,
  runCount,
  createdAt,
  href,
  isLoading,
  error,
}: IStackCard) => {
  if (isLoading) return <Skeleton />

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
      <Text variant="body" className="font-strong">
        Stack
      </Text>
      {status && <Status status={status} variant="badge" />}
      {runCount !== undefined && (
        <Text variant="subtext">
          {runCount} {runCount === 1 ? 'run' : 'runs'}
        </Text>
      )}
      {createdAt && <Time variant="subtext" time={createdAt} format="relative" />}
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
