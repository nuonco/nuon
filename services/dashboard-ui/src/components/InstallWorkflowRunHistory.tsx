'use client'

import React, { type FC, useEffect } from 'react'
import { Timeline } from '@/components/Timeline'
import { revalidateInstallWorkflowHistory } from '@/components/workflow-actions'
import type { TActionWorkflow, TInstallActionWorkflowRun } from '@/types'
import { SHORT_POLL_DURATION } from '@/utils'

interface IInstallWorkflowRunHistory {
  appWorkflows: Array<TActionWorkflow>
  installId: string
  installWorkflowRuns: Array<TInstallActionWorkflowRun>
  orgId: string
  shouldPoll?: boolean
}

export const InstallWorkflowRunHistory: FC<IInstallWorkflowRunHistory> = ({
  appWorkflows,
  installId,
  installWorkflowRuns,
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

  return (
    <Timeline
      emptyMessage="No action workflow runs have happened"
      events={installWorkflowRuns?.map((workflowRun, i) => ({
        id: workflowRun.id,
        status: workflowRun.status,
        underline: (
          <>
            <span>
              {
                appWorkflows?.find((aw) =>
                  aw.configs.find(
                    (awCfg) =>
                      awCfg.id === workflowRun.action_workflow_config_id
                  )
                )?.name
              }
            </span>{' '}
            /
            <span className="!inline truncate max-w-[100px]">
              {workflowRun.trigger_type}
            </span>
          </>
        ),
        time: workflowRun.updated_at,
        href: `/${orgId}/installs/${installId}/actions/${workflowRun.id}`,
        isMostRecent: i === 0,
      }))}
    />
  )
}
