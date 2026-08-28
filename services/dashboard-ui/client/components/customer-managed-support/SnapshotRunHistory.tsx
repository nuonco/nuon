import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { Timeline } from '@/components/common/Timeline'
import { WorkflowTimelineItem } from '@/components/workflows/WorkflowTimeline/WorkflowTimelineItem'
import type { TCustomerManagedSnapshotRun } from '@/lib/ctl-api/installs/customer-managed-support-snapshots'

type CapturedRun = TCustomerManagedSnapshotRun & { created_at: string }

export const CustomerManagedSnapshotRunHistory = ({
  runs,
  kind,
}: {
  runs: TCustomerManagedSnapshotRun[]
  kind: string
}) => {
  const events: CapturedRun[] = runs.map((run) => ({
    ...run,
    created_at: run.started_at,
  }))

  if (!events.length) {
    return (
      <EmptyState
        variant="table"
        emptyTitle={`No ${kind.toLowerCase()} runs captured`}
        emptyMessage={`Run history will appear after this ${kind.toLowerCase()} runs.`}
      />
    )
  }

  return (
    <Timeline<CapturedRun>
      events={events}
      getEventKey={(run) => run.run_id}
      pagination={{ hasNext: false, offset: 0, limit: events.length }}
      renderEvent={(run) => (
        <WorkflowTimelineItem
          id={run.run_id}
          title={run.ref_name}
          status={run.status === 'finished' ? 'success' : run.status}
          createdAt={run.started_at}
          finishedAt={run.finished_at}
          finished={!!run.finished_at}
          createdBy={run.source}
          titleBadges={
            <Badge size="sm" theme="neutral">
              {kind}
            </Badge>
          }
        />
      )}
    />
  )
}
