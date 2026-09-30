import type { ReactNode } from 'react'
import { Avatar } from '@/components/common/Avatar'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TRunSource } from './run-source'

export interface IRunCommit {
  message?: string
  author?: string
  avatarUrl?: string
  sha?: string
  shaUrl?: string
  createdAt?: string
}

export interface IRunSourceCard {
  source: TRunSource
  title: string
  sha?: string
  shaUrl?: string
  author?: string
  status: string
  commit?: IRunCommit
}

const TRIGGER_LABEL: Record<TRunSource['kind'], string> = {
  'pull-request': 'Pull request',
  tag: 'Tag',
  manual: 'Manual run',
  commit: 'Push',
}

const MaybeLink = ({ href, children }: { href?: string; children: ReactNode }) =>
  href ? (
    <Link href={href} isExternal>
      {children}
    </Link>
  ) : (
    <>{children}</>
  )

const SourceIdentity = ({ source }: { source: TRunSource }) => {
  if (source.kind === 'pull-request') {
    return (
      <span className="flex flex-wrap items-center gap-2">
        <Text variant="body" weight="strong" flex>
          <Icon variant="GitPullRequestIcon" />
          <MaybeLink href={source.url}>Pull request #{source.number}</MaybeLink>
        </Text>
        {source.baseBranch ? (
          <Text variant="subtext" theme="neutral">
            into{' '}
            <Text as="span" variant="subtext" family="mono">
              {source.baseBranch}
            </Text>
          </Text>
        ) : null}
        {source.label ? (
          <Badge size="sm" theme="neutral">
            <Icon variant="TagIcon" size={12} />
            {source.label}
          </Badge>
        ) : null}
      </span>
    )
  }

  if (source.kind === 'tag') {
    return (
      <Text variant="body" weight="strong" flex>
        <Icon variant="TagIcon" />
        <MaybeLink href={source.url}>
          <Text as="span" variant="body" weight="strong" family="mono">
            {source.tag}
          </Text>
        </MaybeLink>
      </Text>
    )
  }

  return null
}

export const RunSourceCard = ({
  source,
  title,
  sha,
  shaUrl,
  author,
  status,
  commit,
}: IRunSourceCard) => {
  const commitSha = commit?.sha ?? sha
  const commitUrl = commit?.shaUrl ?? shaUrl
  const commitAuthor = commit?.author ?? author
  const message = commit?.message
  const hasIdentity = source.kind === 'pull-request' || source.kind === 'tag'

  return (
    <section className="border rounded-xl bg-white dark:bg-dark-grey-900 shadow-sm overflow-hidden min-w-0">
      <header className="flex items-center justify-between gap-3 px-5 py-4">
        <Text variant="h3" weight="strong">
          Run information
        </Text>
        <Status status={status} />
      </header>
      <div className="flex flex-col gap-3 p-5 border-t">
        {hasIdentity ? <SourceIdentity source={source} /> : null}
        <Text
          variant="body"
          weight={hasIdentity ? undefined : 'strong'}
          className="break-words"
        >
          {message || title}
        </Text>
        {commitSha || commitAuthor || commit?.createdAt ? (
          <span className="flex flex-wrap items-center gap-x-2 gap-y-1">
            {commitSha ? (
              <Text variant="subtext" family="mono" theme="neutral" flex>
                <Icon variant="GitCommitIcon" />
                <MaybeLink href={commitUrl}>{commitSha.slice(0, 7)}</MaybeLink>
              </Text>
            ) : null}
            {commit?.avatarUrl ? (
              <Avatar
                src={commit.avatarUrl}
                alt={commitAuthor ?? ''}
                size="xs"
                shape="circle"
              />
            ) : null}
            {commitAuthor ? (
              <Text variant="subtext" theme="neutral">
                {commitAuthor}
              </Text>
            ) : null}
            {commit?.createdAt ? (
              <Time
                variant="subtext"
                theme="neutral"
                time={commit.createdAt}
                format="relative"
              />
            ) : null}
          </span>
        ) : null}
        <span className="flex flex-wrap items-center gap-x-2 gap-y-1">
          <Text variant="label" theme="neutral">
            Triggered by
          </Text>
          <Text variant="subtext">{TRIGGER_LABEL[source.kind]}</Text>
        </span>
      </div>
    </section>
  )
}
