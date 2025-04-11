'use client'

import React, { type FC, useEffect } from 'react'
import { Empty } from '@/components/Empty'
import { Timeline } from '@/components/Timeline'
import { Text } from '@/components/Typography'
import { revalidateInstallWorkflowHistory } from '@/components/workflow-actions'
import type { TInstallActionWorkflow } from '@/types'
import { SHORT_POLL_DURATION, humandReadableTriggeredBy } from '@/utils'

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
          emptyTitle="No workflow runs yet"
          emptyMessage={`Waiting on ${actionsWithRecentRuns?.action_workflow?.name} workflow to run.`}
          variant="history"
        />
      }
      events={runs?.map((run, i) => ({
        id: run.id,
        status: run.status,
        underline: (
          <div>
            <Text>
              <span>{action_workflow.name}</span> /
              <span className="!inline truncate max-w-[100px]">
                {humandReadableTriggeredBy(run?.triggered_by_type)}
              </span>
            </Text>
            {run?.created_by ? (
              <Text className="text-cool-grey-600 dark:text-white/70 !text-[10px]">
                Run by: {run?.created_by?.email}
              </Text>
            ) : null}
          </div>
        ),
        time: run.updated_at,
        href: `/${orgId}/installs/${installId}/actions/${action_workflow?.id}/${run.id}`,
        isMostRecent: i === 0,
      }))}
    />
  )
}
