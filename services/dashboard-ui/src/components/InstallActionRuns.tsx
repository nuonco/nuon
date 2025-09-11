'use client'

import { ActionTriggerType } from '@/components/ActionTriggerType'
import { Empty } from '@/components/Empty'
import { Timeline } from '@/components/Timeline'
import { Text } from '@/components/Typography'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useQueryParams } from '@/hooks/use-query-params'
import type { TInstallAction } from '@/types'

interface IInstallActionRuns extends IPollingProps {
  initInstallAction: TInstallAction
  pagination: {
    offset: string
    limit: string
  }
}

export const InstallActionRuns = ({
  initInstallAction,
  pagination,
  pollInterval = 6000,
  shouldPoll = false,
}: IInstallActionRuns) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const params = useQueryParams({
    offset: pagination.offset,
    limit: pagination.limit,
  })
  const { data: installAction } = usePolling<TInstallAction>({
    dependencies: [params],
    initData: initInstallAction,
    path: `/api/orgs/${org.id}/installs/${install.id}/actions/${initInstallAction.action_workflow_id}${params}`,
    pollInterval,
    shouldPoll,
  })

  return (
    <Timeline
      emptyContent={
        <Empty
          emptyTitle="No action runs yet"
          emptyMessage={`Waiting on ${installAction?.action_workflow?.name} action to run.`}
          variant="history"
        />
      }
      events={installAction?.runs?.map((run, i) => ({
        id: run.id,
        status: run?.status_v2?.status,
        underline: (
          <div>
            <span className="flex items-center gap-2">
              <Text variant="reg-12">
                {installAction?.action_workflow.name}
              </Text>{' '}
              /
              <ActionTriggerType
                triggerType={run?.triggered_by_type}
                componentName={run?.run_env_vars?.COMPONENT_NAME}
                componentPath={`/${org.id}/installs/${install.id}/components/${run?.run_env_vars?.COMPONENT_ID}`}
              />
            </span>
            {run?.created_by ? (
              <Text className="!text-[10px]" isMuted>
                Run by: {run?.created_by?.email}
              </Text>
            ) : null}
          </div>
        ),
        time: run.updated_at,
        href:
          run?.status_v2?.status &&
          (run?.status_v2?.status as string) !== 'queued'
            ? `/${org.id}/installs/${install.id}/actions/${installAction?.action_workflow?.id}/${run.id}`
            : null,
        isMostRecent: i === 0,
      }))}
    />
  )
}
