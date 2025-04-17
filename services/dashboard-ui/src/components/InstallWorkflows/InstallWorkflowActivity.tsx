'use client'

import { usePathname } from 'next/navigation'
import React, { type FC, useEffect } from 'react'
import { revalidateData } from '@/components/actions'
import { Text } from '@/components/Typography'
import type { TInstallWorkflow } from '@/types'
import { POLL_DURATION } from '@/utils'
import { YAStatus } from './InstallWorkflowHistory'

interface IInstallWorkflowActivity {
  installWorkflow: TInstallWorkflow
  shouldPoll?: boolean
  pollDuration?: number
}

export const InstallWorkflowActivity: FC<IInstallWorkflowActivity> = ({
  installWorkflow,
  shouldPoll = false,
  pollDuration = POLL_DURATION,
}) => {
  const path = usePathname()

  useEffect(() => {
    const refreshData = () => {
      revalidateData({ path })
    }
    if (shouldPoll) {
      const pollBuild = setInterval(refreshData, pollDuration)

      return () => clearInterval(pollBuild)
    }
  }, [installWorkflow, shouldPoll])

  return (
    <div className="border p-4 rounded-md flex flex-col gap-6">
      <span className="flex w-full justify-between flex-wrap">
        <Text variant="med-14">Install Activity</Text>

        <Text
          variant="reg-12"
          className="text-cool-grey-600 dark:text-white/70"
        >
          {
            installWorkflow?.steps?.filter(
              (s) =>
                s?.status?.status === 'success' ||
                s?.status?.status === 'active'
            ).length
          }{' '}
          of {installWorkflow?.steps?.length} steps completed
        </Text>
      </span>

      <div className="">
        {installWorkflow?.status ? (
          <span className="flex gap-2">
            <YAStatus status={installWorkflow?.status?.status} />
            <Text variant="reg-12">
              {installWorkflow?.status?.status_human_description ||
                'Waiting on workflow to run.'}
            </Text>
          </span>
        ) : null}
      </div>
    </div>
  )
}
