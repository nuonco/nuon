'use client'

import React, { type FC, useEffect } from 'react'
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
            <span className="flex items-center gap-2">
              <Text variant="reg-12">{action_workflow.name}</Text> /
              <Badge className="!inline" variant="code">
                {run?.triggered_by_type}
              </Badge>
            </span>
            {run?.created_by ? (
              <Text className="!text-[10px]" isMuted>
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
