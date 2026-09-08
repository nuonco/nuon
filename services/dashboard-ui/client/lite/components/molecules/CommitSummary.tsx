import type { TVCSCommit } from '@/types'
import { Icon } from '../atoms/Icon'
import { Text } from '../atoms/Text'
import { Time } from './Time'

export interface ICommitSummary {
  commit?: TVCSCommit
  updatedAt?: string
  isLoading?: boolean
}

export const CommitSummary = ({
  commit,
  updatedAt,
  isLoading = false,
}: ICommitSummary) => {
  if (isLoading) {
    return (
      <>
        <Text loading loadingWidth={7} family="mono" />
        <Text loading loadingWidth={20} variant="caption" />
        <Time loading loadingWidth={12} />
      </>
    )
  }

  if (!commit) {
    return (
      <>
        <Text color="tertiary">No commits yet</Text>
        <Text variant="caption" color="tertiary">
          The latest synced commit will appear here.
        </Text>
      </>
    )
  }

  return (
    <>
      <span className="flex items-center gap-2">
        <Icon variant="GitCommitIcon" size={18} />
        <Text family="mono" weight="medium">
          {commit?.sha?.slice(0, 7) ?? '—'}
        </Text>
      </span>
      <Text variant="caption" color="secondary" lines={1}>
        {commit?.message?.split('\n')[0] ?? 'No commit message'}
      </Text>
      <span className="flex flex-wrap items-center gap-x-2 gap-y-1">
        {commit?.author_name ? (
          <Text variant="caption" color="tertiary">
            {commit.author_name}
          </Text>
        ) : null}
        <Time value={commit?.created_at ?? updatedAt} format="relative" />
      </span>
    </>
  )
}
