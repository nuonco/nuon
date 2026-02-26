import { useQuery } from '@tanstack/react-query'
import { ActionTriggerType } from '@/components/actions/ActionTriggerType'
import { Badge } from '@/components/common/Badge'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Timeline, type ITimeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { Text } from '@/components/common/Text'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useQueryParams } from '@/hooks/use-query-params'
import { getInstallAction } from '@/lib'
import type {
  TInstallAction,
  TInstallActionRun,
  TActionConfigTriggerType,
} from '@/types'

interface IInstallActionRunTimeline
  extends Omit<ITimeline<TInstallAction>, 'events' | 'renderEvent'> {
  initInstallAction: TInstallAction
  pollInterval?: number
  shouldPoll?: boolean
}

export const InstallActionRunTimeline = ({
  initInstallAction,
  pagination,
  pollInterval = 20000,
  shouldPoll = false,
}: IInstallActionRunTimeline) => {
  const { install } = useInstall()
  const { org } = useOrg()

  const queryParams = useQueryParams({
    offset: pagination.offset,
    limit: 10,
  })
  const { data: action } = useQuery<TInstallAction>({
    queryKey: ['install-action', org?.id, install?.id, initInstallAction?.action_workflow_id, queryParams],
    queryFn: () =>
      getInstallAction({
        orgId: org.id,
        installId: install.id,
        actionId: initInstallAction?.action_workflow_id,
        limit: 10,
        offset: pagination?.offset,
      }).then((r) => r.data ?? initInstallAction),
    initialData: initInstallAction,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id,
  })

  return (
    <Timeline<TInstallActionRun>
      events={action?.runs}
      pagination={pagination}
      renderEvent={(run) => {
        return (
          <TimelineEvent
            key={run.id}
            caption={<ID>{run?.id}</ID>}
            createdAt={run?.created_at}
            status={run?.status}
            title={
              <span className="flex items-center gap-2">
                <Link
                  href={`/${org.id}/installs/${install.id}/actions/${action?.action_workflow_id}/${run.id}`}
                >
                  {action?.action_workflow?.name} run
                </Link>
                <ActionTriggerType
                  triggerType={
                    run?.triggered_by_type as TActionConfigTriggerType
                  }
                  componentName={run?.run_env_vars?.COMPONENT_NAME}
                  componentPath={`/${org.id}/installs/${install.id}/components/${run?.run_env_vars?.COMPONENT_ID}`}
                  size="sm"
                />
                {run?.status_v2?.status === 'drifted' ? (
                  <Badge variant="code" size="sm">
                    drift scan
                  </Badge>
                ) : null}
              </span>
            }
            underline={
              <Text variant="label" theme="neutral">
                Run by: {run?.created_by?.email}
              </Text>
            }
          />
        )
      }}
    />
  )
}
