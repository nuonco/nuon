import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { ActionTriggerType } from '@/components/actions/ActionTriggerType'
import { Badge } from '@/components/common/Badge'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { Text } from '@/components/common/Text'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallAction } from '@/lib'
import type { TInstallActionRun, TActionConfigTriggerType } from '@/types'

const LIMIT = 10

interface IInstallActionRunTimeline {
  actionId: string
  actionName: string
  pollInterval?: number
  shouldPoll?: boolean
}

export const InstallActionRunTimeline = ({
  actionId,
  actionName,
  pollInterval = 20000,
  shouldPoll = false,
}: IInstallActionRunTimeline) => {
  const { install } = useInstall()
  const { org } = useOrg()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: action } = useQuery({
    queryKey: ['install-action', org?.id, install?.id, actionId, offset],
    queryFn: () =>
      getInstallAction({
        orgId: org.id,
        installId: install.id,
        actionId,
        limit: LIMIT,
        offset,
      }),
    refetchOnMount: 'always',
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id && !!actionId,
  })

  const runs = action?.runs ?? []
  const pagination = { hasNext: runs.length >= LIMIT, offset, limit: LIMIT }

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
                href={`/${org.id}/installs/${install.id}/actions/${actionId}/runs/${run.id}`}
              >
                {actionName} run
              </Link>
              <ActionTriggerType
                triggerType={run?.triggered_by_type as TActionConfigTriggerType}
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
      )}
    />
  )
}
