import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'

export interface ICommitRange {
  message?: string
  author?: string
  sha?: string
  previousSha?: string
  createdAt?: string
}

const shortSha = (sha?: string) => (sha ? sha.slice(0, 7) : undefined)

export const commitRangeLabel = (commit: ICommitRange) => {
  const next = shortSha(commit.sha)
  const previous = shortSha(commit.previousSha)
  if (!next) return previous
  if (!previous || previous === next || commit.previousSha === commit.sha) {
    return next
  }
  return `${previous} → ${next}`
}

export const CommitRange = ({ commit }: { commit: ICommitRange }) => {
  const label = commitRangeLabel(commit)
  if (!label) return null

  return (
    <Tooltip
      position="top"
      tipContentClassName="whitespace-normal max-w-80"
      tipContent={
        <span className="flex flex-col gap-1.5">
          {commit.message ? (
            <Text variant="subtext" className="line-clamp-3 whitespace-normal">
              {commit.message}
            </Text>
          ) : null}
          <span className="flex flex-wrap items-center gap-x-2 gap-y-1">
            <Text variant="subtext" family="mono" theme="neutral">
              {label}
            </Text>
            {commit.author ? (
              <Text variant="subtext" theme="neutral">
                {commit.author}
              </Text>
            ) : null}
            {commit.createdAt ? (
              <Time
                variant="subtext"
                theme="neutral"
                time={commit.createdAt}
                format="relative"
              />
            ) : null}
          </span>
        </span>
      }
    >
      <Text variant="subtext" theme="neutral" family="mono" flex>
        <Icon variant="GitCommitIcon" />
        {label}
      </Text>
    </Tooltip>
  )
}
