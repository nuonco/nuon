'use client'

import React, { type FC, useEffect } from 'react'
import { ActionTriggerType } from '@/components/ActionTriggerType'
import { Badge } from '@/components/Badge'
import { Empty } from '@/components/Empty'
import { Timeline } from '@/components/Timeline'
import { Text } from '@/components/Typography'
import { revalidateInstallWorkflowHistory } from '@/components/workflow-actions'
import type { TInstallActionWorkflow } from '@/types'
import { SHORT_POLL_DURATION } from '@/utils'

interface IInstallWorkflowRunHistory {
  actionsWithRecentRuns: TInstallActionWorkflow
  installId: string
  orgId: string
  shouldPoll?: boolean
}

export const InstallWorkflowRunHistory: FC<IInstallWorkflowRunHistory> = ({
  actionsWithRecentRuns,
  installId,
  orgId,
  shouldPoll = false,
}) => {
  useEffect(() => {
    const revalidateHistory = () => {
      revalidateInstallWorkflowHistory(orgId, installId)
    }

    if (shouldPoll) {
      const pollWorkflowRuns = setInterval(
        revalidateHistory,
        SHORT_POLL_DURATION
      )
      return () => clearInterval(pollWorkflowRuns)
    }
  }, [shouldPoll])

  const { action_workflow, runs } = actionsWithRecentRuns

  return (
    <Timeline
      emptyContent={
        <Empty
          emptyTitle="No action runs yet"
          emptyMessage={`Waiting on ${actionsWithRecentRuns?.action_workflow?.name} action to run.`}
          variant="history"
        />
      }
      events={runs?.map((run, i) => ({
        id: run.id,
        status: run?.status_v2?.status,
        underline: (
          <div>
            <span className="flex items-center gap-2">
              <Text variant="reg-12">{action_workflow.name}</Text> /
              <ActionTriggerType
                triggerType={run?.triggered_by_type}
                componentName={run?.run_env_vars?.COMPONENT_NAME}
                componentPath={`/${orgId}/installs/${installId}/components/${run?.run_env_vars?.COMPONENT_ID}`}
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
            ? `/${orgId}/installs/${installId}/actions/${action_workflow?.id}/${run.id}`
            : null,
        isMostRecent: i === 0,
      }))}
    />
  )
}
