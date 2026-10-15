import { DateTime } from 'luxon'
import {
  CommitLink,
  PullRequestLink,
  TagLink,
} from '@/components/common/GitReferenceLink'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import type { TGroupApprovedVersion } from './group-approved-version'

const detailLine = (version: TGroupApprovedVersion) => {
  const message = version.message?.split('\n')[0]?.trim()
  const clipped =
    message && message.length > 80 ? `${message.slice(0, 77)}...` : message
  const when = version.createdAt
    ? DateTime.fromISO(version.createdAt).toRelative()
    : undefined
  return [clipped, version.author, when].filter(Boolean).join(' · ')
}

const SourceMark = ({ version }: { version: TGroupApprovedVersion }) => {
  if (version.source.kind === 'pull-request') {
    return (
      <PullRequestLink
        number={version.source.number}
        href={version.source.url}
        textVariant="subtext"
      />
    )
  }
  if (version.source.kind === 'tag') {
    return (
      <TagLink
        tag={version.source.tag}
        href={version.source.url}
        textVariant="subtext"
        className="max-w-36 truncate"
      />
    )
  }
  if (version.sha) {
    return (
      <CommitLink sha={version.sha} href={version.shaUrl} textVariant="subtext" />
    )
  }
  if (version.source.kind === 'manual') {
    return (
      <Text variant="subtext" theme="neutral">
        Manual
      </Text>
    )
  }
  return null
}

export const GroupApprovedVersion = ({
  version,
  isLoading,
}: {
  version?: TGroupApprovedVersion
  isLoading?: boolean
}) => {
  if (isLoading) {
    return <Text variant="subtext" loading loadingWidth={16} />
  }
  if (!version) {
    return (
      <Text variant="subtext" theme="neutral">
        None
      </Text>
    )
  }

  const mark = (
    <span className="flex min-w-0 items-center gap-2">
      {version.version != null ? (
        <Text variant="subtext" family="mono" className="shrink-0">
          v{version.version}
        </Text>
      ) : null}
      <SourceMark version={version} />
    </span>
  )
  const detail = detailLine(version)
  if (!detail) return mark

  return (
    <Tooltip position="top" tipContent={detail}>
      {mark}
    </Tooltip>
  )
}
