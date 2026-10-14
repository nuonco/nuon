import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import {
  getRunTitle,
  getRunTrigger,
} from '@/components/branches/shared/run-title'
import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { LabeledValue } from '@/components/common/LabeledValue'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import type { TInstallOverviewCommit } from '@/types'
import { manualRunPinLabel } from '@/utils/branch-utils'
import { getWorkflowBadge } from '@/utils/workflow-utils'

export interface ISuccessfulBranchRuns {
  branchName?: string
  runs: TInstallOverviewCommit[]
  runsBehind?: number | null
  getRunHref: (run?: TInstallOverviewCommit) => string | undefined
  repo?: string
  selectedCommit?: TInstallOverviewCommit
}

export const SuccessfulBranchRuns = ({
  branchName,
  runs,
  runsBehind,
  getRunHref,
  repo,
  selectedCommit,
}: ISuccessfulBranchRuns) => {
  const hidden = (runsBehind ?? runs.length) - runs.length
  const selectedHref = getRunHref(selectedCommit)

  return (
    <div className="flex flex-col gap-4" data-testid="successful-runs">
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
        These successful app branch runs have configurations newer than this
        install’s selected configuration. Failed, canceled, preview, and
        no-change runs don’t count. Newest runs are shown first.
      </Text>
      {runs.length === 0 ? (
        <EmptyState
          variant="history"
          emptyTitle="No successful runs"
          emptyMessage="Newer configurations will appear here after an app branch run succeeds."
        />
      ) : (
        <Timeline
          events={runs}
          groupByDate={false}
          getEventKey={(run) => run.run_id ?? run.workflow_id ?? run.sha}
          pagination={{ hasNext: false, offset: 0, limit: 50 }}
          renderEvent={(run) => {
            const href = getRunHref(run)
            const branchRun = {
              metadata: run.metadata,
              head_sha: run.sha,
              vcs_connection_commit: { message: run.message },
            }
            const title = getRunTitle({ app_branch_runs: [branchRun] })
            const trigger = getRunTrigger(branchRun)
            const pinLabel = manualRunPinLabel(branchRun)
            return (
              <TimelineEvent
                data-testid="successful-run"
                status={run.run_status ?? 'success'}
                createdAt={run.created_at ?? ''}
                title={
                  href ? (
                    <span className="flex items-center gap-4 mb-1">
                      <Link variant="inline" href={href}>
                        {title}
                      </Link>
                    </span>
                  ) : (
                    title
                  )
                }
                badge={getWorkflowBadge({
                  status: { status: 'success' },
                })}
                additionalCaption={
                  trigger === 'manual' ? (
                    <>
                      <Badge variant="code" size="sm">
                        manual
                      </Badge>
                      {pinLabel ? (
                        <Badge variant="code" size="sm" theme="info">
                          {pinLabel}
                        </Badge>
                      ) : null}
                    </>
                  ) : trigger === 'github_label' &&
                    run.metadata?.github_label ? (
                    <LabelBadge
                      labelKey="label"
                      labelValue={run.metadata.github_label}
                      size="sm"
                    />
                  ) : undefined
                }
                underline={
                  <span className="flex flex-col gap-4 mt-2">
                    <BranchRunCommit
                      href={href}
                      sha={run.sha}
                      message={run.message || 'App configuration update'}
                      author={run.author}
                      createdAt={run.created_at}
                      repo={repo}
                      className="ml-2 border-l pl-4 py-1"
                      showStatus={false}
                    />
                  </span>
                }
              />
            )
          }}
        />
      )}
      {hidden > 0 ? (
        <Text variant="subtext" theme="neutral">
          Showing the newest {runs.length} of {runsBehind} successful app branch
          runs on {branchName ?? 'this branch'}.
        </Text>
      ) : null}
    </div>
  )
}
