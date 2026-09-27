import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TRunSource } from './run-source'

export interface IRunSourceCard {
  source: TRunSource
  title: string
  sha?: string
  shaUrl?: string
  author?: string
  status: string
}

const SOURCE_LABEL: Record<TRunSource['kind'], string> = {
  'pull-request': 'Pull request',
  tag: 'Tag',
  commit: 'Commit',
}

const SourceIdentity = ({ source }: { source: TRunSource }) => {
  if (source.kind === 'pull-request') {
    const number = `#${source.number}`
    return (
      <span className="flex flex-wrap items-center gap-2">
        <Text variant="subtext" weight="strong" flex>
          <Icon variant="GitPullRequestIcon" />
          {source.url ? (
            <Link href={source.url} isExternal>
              {number}
            </Link>
          ) : (
            number
          )}
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
      <Text variant="subtext" family="mono" weight="strong" flex>
        <Icon variant="TagIcon" />
        {source.url ? (
          <Link href={source.url} isExternal>
            {source.tag}
          </Link>
        ) : (
          source.tag
        )}
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
}: IRunSourceCard) => {
  const shortSha = sha?.slice(0, 7)

  return (
    <Card className="!p-4 !gap-3 min-w-0">
      <span className="flex items-center justify-between gap-2">
        <Text variant="label" theme="neutral">
          {SOURCE_LABEL[source.kind]}
        </Text>
        <Status status={status} />
      </span>
      <SourceIdentity source={source} />
      <Text variant="body" weight="strong">
        {title}
      </Text>
      <span className="flex flex-wrap items-center gap-x-3 gap-y-1">
        {shortSha ? (
          <Text variant="subtext" family="mono" theme="neutral" flex>
            <Icon variant="GitCommitIcon" />
            {shaUrl ? (
              <Link href={shaUrl} isExternal>
                {shortSha}
              </Link>
            ) : (
              shortSha
            )}
          </Text>
        ) : null}
        {author ? (
          <Text variant="subtext" family="mono" theme="neutral">
            {author}
          </Text>
        ) : null}
      </span>
    </Card>
  )
}
