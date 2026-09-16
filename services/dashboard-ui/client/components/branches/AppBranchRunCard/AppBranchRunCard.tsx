import { Card } from '@/components/common/Card'
import { LabeledValue } from '@/components/common/LabeledValue'
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
  sourceRepo?: string
}

export const AppBranchRunCard = ({
  appId,
  buildStatus,
  orgId,
  run,
  sourceCommit,
  sourceHref,
  sourceRepo,
}: IAppBranchRunCard) => {
  const runCommit = run?.vcs_connection_commit
  const branchId = run?.app_branch?.id
  const runHref =
    orgId && appId && branchId && run?.workflow_id
      ? `/${orgId}/apps/${appId}/branches/${branchId}/runs/${run.workflow_id}`
      : undefined
  const sameCommit =
    !!sourceCommit?.sha &&
    !!runCommit?.sha &&
    sourceCommit.sha === runCommit.sha

  if (!sourceCommit && !run) return null

  return (
    <Card className="gap-4">
      <div className="flex items-start justify-between gap-4">
        <div className="flex flex-col gap-1">
          <Text weight="strong">Git source</Text>
          {sourceRepo ? (
            <Text variant="subtext" theme="neutral">
              {sourceRepo}
            </Text>
          ) : null}
        </div>
        {runHref ? <Link href={runHref}>App branch update</Link> : null}
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <LabeledValue label="Source commit">
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
        </LabeledValue>

        {run ? (
          <LabeledValue label="App branch commit">
            {sameCommit ? (
              <Text variant="subtext" theme="neutral">
                Same as app branch since it is in the same repository.
              </Text>
            ) : runCommit ? (
              <BranchRunCommit
                status={run.status}
                href={runHref}
                message={runCommit.message?.split('\n')[0]}
                author={runCommit.author_name}
                avatarUrl={runCommit.author_avatar_url}
                sha={runCommit.sha}
                createdAt={runCommit.created_at ?? run.created_at}
              />
            ) : (
              <Text variant="subtext" theme="neutral">
                No app branch commit recorded
              </Text>
            )}
          </LabeledValue>
        ) : null}
      </div>
    </Card>
  )
}
