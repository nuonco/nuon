import { ActionTriggerType } from '@/components/actions/ActionTriggerType'
import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { Text } from '@/components/common/Text'
import type { TInstallActionRun, TActionConfigTriggerType } from '@/types'

interface IInstallActionRunTimeline {
  actionId: string
  actionName: string
  runs: TInstallActionRun[]
  basePath: string
  pagination: { hasNext?: boolean; offset: number; limit: number }
}

export const InstallActionRunTimeline = ({
  actionId,
  actionName,
  runs,
  basePath,
  pagination,
}: IInstallActionRunTimeline) => {
  if (runs.length === 0 && pagination.offset === 0) {
    return (
      <EmptyState
        className="mt-12"
        variant="history"
        size="sm"
        emptyTitle="No runs yet"
        emptyMessage="This action has not been run yet. Trigger a run to see history here."
      />
    )
  }

  return (
    <Timeline<TInstallActionRun>
      events={runs}
      pagination={pagination}
      renderEvent={(run) => (
        <TimelineEvent
          key={run.id}
          caption={<ID>{run?.id}</ID>}
          createdAt={run?.created_at}
          status={run?.status}
          title={
            <span className="flex items-center gap-2">
              <Link
                href={`${basePath}/actions/${actionId}/runs/${run.id}`}
              >
                {actionName} run
              </Link>
              {run?.status_v2?.status === 'drifted' ? (
                <Badge variant="code" size="sm">
                  drift scan
                </Badge>
              ) : null}
            </span>
          }
          underline={
            <div className="flex flex-col gap-1">
              <ActionTriggerType
                triggerType={run?.triggered_by_type as TActionConfigTriggerType}
                componentName={run?.run_env_vars?.COMPONENT_NAME}
                componentPath={`${basePath}/components/${run?.run_env_vars?.COMPONENT_ID}`}
                size="sm"
              />
              <Text variant="label" theme="neutral">
                Run by: {run?.created_by?.email}
              </Text>
            </div>
          }
        />
      )}
    />
  )
}
