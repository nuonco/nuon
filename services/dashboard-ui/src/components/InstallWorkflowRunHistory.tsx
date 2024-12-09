'use client'

import classNames from 'classnames'
import React, { type FC, useEffect } from 'react'
import { Link } from '@/components/Link'
import { StatusBadge } from '@/components/Status'
import { Time } from '@/components/Time'
import { Text } from '@/components/Typography'
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
      const pollBuilds = setInterval(revalidateHistory, SHORT_POLL_DURATION)
      return () => clearInterval(pollBuilds)
    }
  }, [shouldPoll])

  return (
    <div className="flex flex-col gap-2">
      {installWorkflowRuns.map((w, i) => (
        <Link
          key={w.id}
          className="!block w-full !p-0"
          href={`/${orgId}/installs/${installId}/workflows/${w.id}`}
          variant="ghost"
        >
          <div
            className={classNames('flex items-center justify-between p-4', {
              'border rounded-md shadow-sm': i === 0,
            })}
          >
            <div className="flex flex-col">
              <span className="flex items-center gap-2">
                <StatusBadge
                  status={w.status}
                  isStatusTextHidden
                  isWithoutBorder
                />
              </span>

              <Text className="flex items-center gap-2 ml-3.5" variant="reg-12">
                <span>
                  {
                    appWorkflows?.find((aw) =>
                      aw.configs.find(
                        (awCfg) => awCfg.id === w.action_workflow_config_id
                      )
                    )?.name
                  }
                </span>{' '}
                /
                <span className="!inline truncate max-w-[100px]">
                  {w.trigger_type}
                </span>
              </Text>
            </div>

            <div className="flex items-center gap-2">
              <Time time={w.updated_at} format="relative" variant="reg-12" />
            </div>
          </div>
        </Link>
      ))}
    </div>
  )
}
