import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { Text } from '@/components/common/Text'
import type { TInstallRunbookRun } from '@/lib/ctl-api/installs/runbooks/get-install-runbooks'

interface IRunbookRunTimeline {
  runbookName: string
  runs: TInstallRunbookRun[]
  basePath: string
}

export const RunbookRunTimeline = ({
  runbookName,
  runs,
  basePath,
}: IRunbookRunTimeline) => {
  if (runs.length === 0) {
    return (
      <EmptyState
        className="mt-12"
        variant="history"
        size="sm"
        emptyTitle="No runs yet"
        emptyMessage="This runbook has not been run yet. Trigger a run to see history here."
      />
    )
  }

  return (
    <Timeline<TInstallRunbookRun>
      events={runs}
      pagination={{ hasNext: false, offset: 0, limit: runs.length }}
      renderEvent={(run) => {
        const wfStatus =
          typeof run.install_workflow?.status === 'object'
            ? (run.install_workflow.status as { status?: string })?.status
            : run.install_workflow?.status
        const status = wfStatus ?? run.status ?? 'unknown'
        const workflowId =
          run.install_workflow_id ?? run.install_workflow?.id

        return (
          <TimelineEvent
            key={run.id}
            caption={<ID>{run.id}</ID>}
            createdAt={run.created_at ?? ''}
            status={status}
            title={
              workflowId ? (
                <Link href={`${basePath}/workflows/${workflowId}`}>
                  {runbookName} run
                </Link>
              ) : (
                <span>{runbookName} run</span>
              )
            }
            underline={
              run.created_by?.email ? (
                <Text variant="label" theme="neutral">
                  Run by: {run.created_by.email}
                </Text>
              ) : undefined
            }
          />
        )
      }}
    />
  )
}
