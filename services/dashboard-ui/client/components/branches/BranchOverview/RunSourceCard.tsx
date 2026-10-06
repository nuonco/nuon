import { useState } from 'react'
import { AnimatedHeight } from '@/components/common/AnimatedHeight'
import { Avatar } from '@/components/common/Avatar'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import {
  CommitLink,
  PullRequestLink,
  TagLink,
} from '@/components/common/GitReferenceLink'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { cn } from '@/utils/classnames'
import type { TRunSource } from './run-source'

const COLLAPSED_LINES = 6
const COLLAPSED_CHARS = 480

const messageOverflows = (text: string) =>
  text.split('\n').length > COLLAPSED_LINES || text.length > COLLAPSED_CHARS

const CommitMessage = ({ text }: { text: string }) => {
  const [expanded, setExpanded] = useState(false)
  const overflows = messageOverflows(text)
  const collapsed = overflows && !expanded

  return (
    <div className="flex flex-col items-start gap-1 min-w-0 w-full">
      <AnimatedHeight className="w-full">
        <div
          className={cn(
            collapsed &&
              'max-h-32 overflow-hidden [mask-image:linear-gradient(to_bottom,black_65%,transparent)]'
          )}
        >
          <Text
            as="p"
            variant="subtext"
            className="break-words whitespace-pre-line"
          >
            {text}
          </Text>
        </div>
      </AnimatedHeight>
      {overflows ? (
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="-ml-2"
          aria-expanded={expanded}
          onClick={() => setExpanded((open) => !open)}
        >
          {expanded ? 'Show less' : 'Show more'}
          <Icon
            variant={expanded ? 'CaretUpIcon' : 'CaretDownIcon'}
            size={14}
          />
        </Button>
      ) : null}
    </div>
  )
}

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

const SourceIdentity = ({ source }: { source: TRunSource }) => {
  if (source.kind === 'pull-request') {
    return (
      <span className="flex flex-wrap items-center gap-2">
        <PullRequestLink
          number={source.number}
          href={source.url}
          label={`Pull request #${source.number}`}
          textVariant="body"
          weight="strong"
        />
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
      <TagLink
        tag={source.tag}
        href={source.url}
        textVariant="body"
        weight="strong"
      />
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
        <CommitMessage key={message || title} text={message || title} />
        {commitSha || commitAuthor || commit?.createdAt ? (
          <span className="flex flex-wrap items-center gap-x-2 gap-y-1">
            {commitSha ? (
              <CommitLink sha={commitSha} href={commitUrl} />
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
