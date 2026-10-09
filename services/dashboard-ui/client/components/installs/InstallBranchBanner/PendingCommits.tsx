import { Badge } from '@/components/common/Badge'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { Card } from '@/components/common/Card'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TInstallOverviewCommit } from '@/types'

export interface IPendingCommits {
  branchName?: string
  commits: TInstallOverviewCommit[]
  commitsBehind?: number | null
  hrefFor: (commit?: TInstallOverviewCommit) => string | undefined
  repo?: string
  selectedCommit?: TInstallOverviewCommit
}

export const PendingCommits = ({
  branchName,
  commits,
  commitsBehind,
  hrefFor,
  repo,
  selectedCommit,
}: IPendingCommits) => {
  const hidden = (commitsBehind ?? commits.length) - commits.length
  const selectedHref = hrefFor(selectedCommit)

  return (
    <div className="flex flex-col gap-4" data-testid="pending-commits">
      {selectedCommit?.sha ? (
        <LabeledValue label="Selected commit">
          <BranchRunCommit
            displayVariant="inline"
            href={selectedHref}
            sha={selectedCommit.sha}
            message={selectedCommit.message || 'App configuration update'}
            author={selectedCommit.author}
            createdAt={selectedCommit.created_at}
            repo={repo}
            showStatus={false}
          />
        </LabeledValue>
      ) : null}
      <Text variant="subtext" theme="neutral">
        These app configuration changes are newer than the install’s selected
        commit. View a branch run to review its changes and rollout status.
      </Text>
      {commits.map((commit) => {
        const href = hrefFor(commit)
        return (
          <Card
            key={commit.workflow_id ?? commit.run_id ?? commit.sha}
            className="!p-4 !gap-3 min-w-0"
            data-testid="pending-commit"
          >
            <div className="flex flex-wrap items-start justify-between gap-3">
              <BranchRunCommit
                className="min-w-0"
                sha={commit.sha}
                message={commit.message || 'App configuration update'}
                author={commit.author}
                createdAt={commit.created_at}
                repo={repo}
                showStatus={false}
              />
              {href ? (
                <Link href={href} textVariant="subtext">
                  View branch run
                </Link>
              ) : null}
            </div>
            <div className="flex flex-wrap items-center gap-2">
              {commit.run_status ? <Status status={commit.run_status} /> : null}
              {commit.awaiting_approval ? (
                <Badge size="sm" theme="warn">
                  Awaiting approval
                </Badge>
              ) : null}
            </div>
          </Card>
        )
      })}
      {hidden > 0 ? (
        <Text variant="subtext" theme="neutral">
          Showing the newest {commits.length} of {commitsBehind} commits on{' '}
          {branchName ?? 'this branch'}.
        </Text>
      ) : null}
    </div>
  )
}
