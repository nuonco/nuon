import { ActionTriggerType } from '@/components/actions/ActionTriggerType'
import { Badge } from '@/components/common/Badge'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
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
  return (
    <Timeline<TInstallActionRun>
      events={runs}
      pagination={pagination}
      renderEvent={(run) => (
        <TimelineEvent
          key={run.id}
          caption={<ID>{run?.id}</ID>}
          createdAt={run?.created_at}
          createdBy={run?.created_by?.email}
          status={run?.status}
          title={
            <span className="flex items-center gap-2">
              <Link
                href={`${basePath}/actions/${actionId}/runs/${run.id}`}
              >
                {actionName} run
              </Link>
              <ActionTriggerType
                triggerType={run?.triggered_by_type as TActionConfigTriggerType}
                componentName={run?.run_env_vars?.COMPONENT_NAME}
                componentPath={`${basePath}/components/${run?.run_env_vars?.COMPONENT_ID}`}
              />
              {run?.status_v2?.status === 'drifted' ? (
                <Badge variant="code">
                  drift scan
                </Badge>
              ) : null}
            </span>
          }
        />
      )}
    />
  )
}
