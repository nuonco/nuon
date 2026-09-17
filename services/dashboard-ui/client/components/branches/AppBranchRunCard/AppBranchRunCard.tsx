import { Card } from '@/components/common/Card'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import type { TAppBranchRun, TVCSCommit } from '@/types'

export interface IAppBranchRunCard {
  appId?: string
  buildStatus?: string
  orgId?: string
  run?: TAppBranchRun
  sourceCommit?: TVCSCommit
  sourceHref?: string
}

export const AppBranchRunCard = ({
  appId,
  buildStatus,
  orgId,
  run,
  sourceCommit,
  sourceHref,
}: IAppBranchRunCard) => {
  const branchId = run?.app_branch?.id
  const runHref =
    orgId && appId && branchId && run?.workflow_id
      ? `/${orgId}/apps/${appId}/branches/${branchId}/runs/${run.workflow_id}`
      : undefined

  if (!sourceCommit && !run) return null

  return (
    <Card className="gap-4">
      <div className="flex items-start justify-between gap-4">
        <Text weight="strong">Build source</Text>
        {runHref ? <Link href={runHref}>View App Branch Run</Link> : null}
      </div>

      {sourceCommit ? (
        <BranchRunCommit
          status={buildStatus}
          href={sourceHref}
          message={sourceCommit.message?.split('\n')[0]}
          author={sourceCommit.author_name}
          avatarUrl={sourceCommit.author_avatar_url}
          sha={sourceCommit.sha}
          createdAt={sourceCommit.created_at}
        />
      ) : (
        <Text variant="subtext" theme="neutral">
          No source commit recorded
        </Text>
      )}
    </Card>
  )
}
