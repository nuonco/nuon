import { useSearchParams } from 'react-router'
import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import type { TStatusType } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Timeline } from '@/components/common/Timeline'
import { WorkflowTimelineItem } from '@/components/workflows/WorkflowTimeline/WorkflowTimelineItem'
import { useCustomerManagedSupportSnapshot } from '@/hooks/use-customer-managed-support-snapshot'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { CustomerManagedSnapshotContent } from './SnapshotEmpty'

export const CustomerManagedSnapshotRuns = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { snapshot } = useCustomerManagedSupportSnapshot()
  const [searchParams] = useSearchParams()
  const runs = snapshot?.snapshot.runs ?? []
  const snapshotId = searchParams.get('snapshot')
  const timelineRuns = runs.map((run) => ({
    ...run,
    created_at: run.started_at,
  }))

  return (
    <CustomerManagedSnapshotContent>
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Run history
        </Text>
        <Text variant="subtext" theme="neutral">
          View captured installation, upgrade, action, runbook, drift, and plan
          runs.
        </Text>
      </HeadingGroup>
      {runs.length === 0 ? (
        <EmptyState
          variant="table"
          emptyTitle="No runs captured"
          emptyMessage="Runs will appear after operations execute and the customer captures another snapshot."
        />
      ) : (
        <div className="flex flex-col gap-6">
          <Timeline
            events={timelineRuns}
            getEventKey={(run) => run.run_id}
            pagination={{
              hasNext: false,
              offset: 0,
              limit: timelineRuns.length,
            }}
            renderEvent={(run) => (
              <WorkflowTimelineItem
                id={run.run_id}
                title={run.ref_name || run.ref_id || run.run_id}
                status={run.status as TStatusType}
                createdAt={run.started_at}
                finishedAt={run.finished_at}
                finished={!!run.finished_at}
                href={`/${org.id}/installs/${install.id}/workflows/${run.run_id}${snapshotId ? `?snapshot=${snapshotId}` : ''}`}
                additionalCaption={
                  <span className="flex items-center gap-2">
                    <Badge variant="code" size="sm">
                      {run.ref_kind || 'install'}
                    </Badge>
                    {run.source ? (
                      <Badge variant="code" size="sm">
                        {run.source}
                      </Badge>
                    ) : null}
                  </span>
                }
              />
            )}
          />
        </div>
      )}
    </CustomerManagedSnapshotContent>
  )
}
