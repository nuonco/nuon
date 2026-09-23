import { Card } from '@/components/common/Card'
import { Text } from '@/components/common/Text'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import type { TAppBranchRun } from '@/types'

export interface ITrackedRunCard {
  emptyMessage: string
  href?: string
  label: string
  run?: TAppBranchRun
}

export const TrackedRunCard = ({
  emptyMessage,
  href,
  label,
  run,
}: ITrackedRunCard) => {
  const commit = run?.vcs_connection_commit

  return (
    <Card className="!p-4 !gap-3">
      <Text variant="body" weight="strong">
        {label}
      </Text>
      {run ? (
        <BranchRunCommit
          status={run.status}
          href={href}
          message={commit?.message?.split('\n')[0]}
          author={commit?.author_name}
          avatarUrl={commit?.author_avatar_url}
          sha={commit?.sha ?? run.head_sha}
          createdAt={run.created_at}
        />
      ) : (
        <Text variant="subtext" theme="neutral">
          {emptyMessage}
        </Text>
      )}
    </Card>
  )
}
