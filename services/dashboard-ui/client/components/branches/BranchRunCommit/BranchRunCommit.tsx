import { Avatar } from '@/components/common/Avatar'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { cn } from '@/utils/classnames'

export interface IBranchRunCommit {
  status?: string
  href?: string
  isExternal?: boolean
  message?: string
  author?: string
  avatarUrl?: string
  sha?: string
  createdAt?: string
  className?: string
  showStatus?: boolean
}

export const BranchRunCommit = ({
  status,
  href,
  isExternal,
  message,
  author,
  avatarUrl,
  sha,
  createdAt,
  className,
  showStatus = true,
}: IBranchRunCommit) => {
  const hasMeta = !!author || !!createdAt

  return (
    <div className={cn('flex flex-col gap-1 min-w-0', className)}>
      <div className="flex items-center gap-2 min-w-0">
        {showStatus ? (
          <Status
            status={status ?? 'pending'}
            isWithoutText
            className="shrink-0"
          />
        ) : null}
        {sha ? (
          <Text
            variant="subtext"
            theme="neutral"
            family="mono"
            className="shrink-0"
          >
            {sha.slice(0, 7)}
          </Text>
        ) : null}
        <div className="min-w-0 flex-1 text-[13px] leading-5">
          {href ? (
            <Link
              href={href}
              isExternal={isExternal}
              variant="inline"
              className="block truncate !w-full"
            >
              {message || 'View run'}
            </Link>
          ) : (
            <span className="block truncate text-cool-grey-950 dark:text-white">
              {message || 'Run in progress'}
            </span>
          )}
        </div>
      </div>

      {hasMeta ? (
        <div className="flex items-center gap-1.5 min-w-0">
          {avatarUrl ? (
            <Avatar
              src={avatarUrl}
              alt={author ?? ''}
              size="xs"
              shape="circle"
              className="shrink-0"
            />
          ) : null}
          {author ? (
            <Text
              variant="subtext"
              theme="neutral"
              className="truncate min-w-0"
            >
              {author}
            </Text>
          ) : null}
          {author && createdAt ? (
            <Text variant="subtext" theme="neutral" className="shrink-0">
              ·
            </Text>
          ) : null}
          {createdAt ? (
            <Time
              variant="subtext"
              theme="neutral"
              className="shrink-0"
              time={createdAt}
              format="relative"
            />
          ) : null}
        </div>
      ) : null}
    </div>
  )
}
