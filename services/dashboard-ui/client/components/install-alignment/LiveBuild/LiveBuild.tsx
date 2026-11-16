import {
  CommitLink,
  PullRequestLink,
  TagLink,
} from '@/components/common/GitReferenceLink'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TInstallAlignmentOrigin } from '@/types'
import { humanize } from '@/utils/string-utils'

const SourceMark = ({
  origin,
  repo,
}: {
  origin: TInstallAlignmentOrigin
  repo?: string
}) => {
  const run = origin.run
  if (run?.metadata?.pr_number) {
    return (
      <PullRequestLink
        number={run.metadata.pr_number}
        repo={repo}
        className="shrink-0"
      />
    )
  }
  if (run?.metadata?.tag) {
    return (
      <TagLink
        tag={run.metadata.tag}
        repo={repo}
        className="max-w-36 truncate shrink-0"
      />
    )
  }
  return <CommitLink sha={run?.sha} repo={repo} className="shrink-0" />
}

export const RunRef = ({
  emptyLabel = 'Not deployed',
  href,
  origin,
  repo,
}: {
  emptyLabel?: string
  href?: string
  origin?: TInstallAlignmentOrigin
  repo?: string
}) => {
  if (!origin?.run && !origin?.workflow_type) {
    return (
      <Text as="span" variant="subtext" theme="neutral">
        {emptyLabel}
      </Text>
    )
  }
  const label = origin.run
    ? origin.run.message?.split('\n')[0] || 'View run'
    : humanize(origin.workflow_type!)
  return (
    <span className="flex items-center gap-1.5 min-w-0" data-testid="run-ref">
      {origin.run ? <SourceMark origin={origin} repo={repo} /> : null}
      {href ? (
        <Link href={href} className="truncate min-w-0">
          {label}
        </Link>
      ) : (
        <Text as="span" variant="subtext" className="truncate min-w-0">
          {label}
        </Text>
      )}
    </span>
  )
}

export interface ILiveBuild {
  appliedAt?: string
  deployHref?: string
  origin?: TInstallAlignmentOrigin
  originHref?: string
  repo?: string
  variant?: 'deploy' | 'sync'
}

export const LiveBuild = ({
  appliedAt,
  deployHref,
  origin,
  originHref,
  repo,
  variant = 'deploy',
}: ILiveBuild) => (
  <div
    className="flex flex-wrap items-center gap-x-1.5 gap-y-1"
    data-testid="live-build"
  >
    <Text as="span" variant="subtext" theme="neutral">
      {variant === 'sync' ? 'Synced' : 'Deployed'}
    </Text>
    {origin ? <RunRef origin={origin} href={originHref} repo={repo} /> : null}
    {appliedAt ? (
      <Text as="span" variant="subtext" theme="neutral" flex nowrap>
        {origin ? '·' : null}
        <Time
          time={appliedAt}
          format="relative"
          variant="subtext"
          theme="neutral"
        />
      </Text>
    ) : null}
    {deployHref ? (
      <Link href={deployHref} className="ml-1.5">
        {variant === 'sync' ? 'View sync' : 'View deployment'}
      </Link>
    ) : null}
  </div>
)
