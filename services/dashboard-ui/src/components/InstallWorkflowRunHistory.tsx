'use client'

import React, { type FC, useEffect } from 'react'
import { Timeline } from '@/components/Timeline'
import { revalidateInstallWorkflowHistory } from '@/components/workflow-actions'
import type { TActionWorkflow, TInstallActionWorkflowRun } from '@/types'
import { SHORT_POLL_DURATION, humandReadableTriggeredBy } from '@/utils'

interface IInstallWorkflowRunHistory {
  actionsWithRecentRuns: {
    action_workflow: TActionWorkflow
    recent_runs: Array<TInstallActionWorkflowRun>
  }
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

  const { action_workflow, recent_runs } = actionsWithRecentRuns

  return (
    <Timeline
      emptyMessage="No action workflow runs have happened"
      events={recent_runs?.map((run, i) => ({
        id: run.id,
        status: run.status,
        underline: (
          <>
            <span>{action_workflow.name}</span> /
            <span className="!inline truncate max-w-[100px]">
              {humandReadableTriggeredBy(run?.triggered_by_type)}
            </span>
          </>
        ),
        time: run.updated_at,
        href: `/${orgId}/installs/${installId}/actions/${action_workflow?.id}/${run.id}`,
        isMostRecent: i === 0,
      }))}
    />
  )
}


