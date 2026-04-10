import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'

export interface IRunnerCard {
  status?: string
  href?: string
  isLoading?: boolean
  error?: string
}

const Skeleton = () => (
  <div className="flex w-fit items-center gap-3 rounded-lg border border-cool-grey-200 dark:border-cool-grey-800 px-3 py-2.5 animate-pulse">
    <div className="h-4 w-16 rounded bg-cool-grey-200 dark:bg-cool-grey-800" />
    <div className="h-4 w-14 rounded bg-cool-grey-200 dark:bg-cool-grey-800" />
  </div>
)

export const RunnerCard = ({
  status,
  href,
  isLoading,
  error,
}: IRunnerCard) => {
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
        Runner
      </Text>
      {status && <Status status={status} variant="badge" />}
    </div>
  )

  if (href) {
    return <Link href={href} variant="ghost" className="flex !p-0 no-underline">{content}</Link>
  }

  return content
}
