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

export const StackCard = ({
  status,
  runCount,
  createdAt,
  href,
  isLoading,
  error,
}: IStackCard) => {
  if (isLoading) {
    return (
      <div className="flex w-fit items-center gap-3 rounded-lg border px-3 py-2.5">
        <Text variant="body" className="font-strong">
          Stack
        </Text>
        <Status loading variant="badge" />
        <Text variant="subtext" loading loadingWidth={6} />
        <Text variant="subtext" loading loadingWidth={10} />
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
      <Text variant="body" className="font-strong">
        Stack
      </Text>
      {status && <Status status={status} variant="badge" />}
      {runCount !== undefined && (
        <Text variant="subtext">
          {runCount} {runCount === 1 ? 'run' : 'runs'}
        </Text>
      )}
      {createdAt && (
        <Time variant="subtext" time={createdAt} format="relative" />
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
