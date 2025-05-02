'use client'

import { usePathname } from 'next/navigation'
import React, { type FC, useEffect } from 'react'
import { revalidateData } from '@/components/actions'
import { Text } from '@/components/Typography'
import type { TInstallWorkflow } from '@/types'
import { SHORT_POLL_DURATION, sentanceCase } from '@/utils'
import { YAStatus } from './InstallWorkflowHistory'

interface IInstallWorkflowActivity {
  installWorkflow: TInstallWorkflow
  shouldPoll?: boolean
  pollDuration?: number
}

export const InstallWorkflowActivity: FC<IInstallWorkflowActivity> = ({
  installWorkflow,
  shouldPoll = false,
  pollDuration = SHORT_POLL_DURATION,
}) => {
  const path = usePathname()

  useEffect(() => {
    const refreshData = () => {
      revalidateData({ path })
    }
    if (shouldPoll) {
      const pollBuild = setInterval(refreshData, pollDuration)

      if (installWorkflow?.finished) {
        clearInterval(pollBuild)
      }

      return () => clearInterval(pollBuild)
    }
  }, [installWorkflow, shouldPoll])

  return (
    <div className="">
      <span className="flex w-full justify-between flex-wrap">
        <span className="flex flex-col gap-0">
          <span className="flex items-center gap-4">
            <span>&#x1F680;</span>
            <progress
              className="rounded-lg [&::-webkit-progress-bar]:rounded-lg [&::-webkit-progress-value]:rounded-lg   [&::-webkit-progress-bar]:bg-cool-grey-300 [&::-webkit-progress-value]:bg-green-400 [&::-moz-progress-bar]:bg-green-400 [&::-webkit-progress-value]:transition-all [&::-webkit-progress-value]:duration-500 [&::-moz-progress-bar]:transition-all [&::-moz-progress-bar]:duration-500 h-[8px]"
              max={installWorkflow?.steps?.length}
              value={
                installWorkflow?.steps?.filter(
                  (s) =>
                    s?.status?.status === 'success' ||
                    s?.status?.status === 'active' ||
                    s?.status?.status === 'error'
                ).length
              }
            />
          </span>

          <Text
            variant="reg-12"
            className="text-cool-grey-600 dark:text-white/70 self-end"
          >
            {
              installWorkflow?.steps?.filter(
                (s) =>
                  s?.status?.status === 'success' ||
                  s?.status?.status === 'active' ||
                  s?.status?.status === 'error'
              ).length
            }{' '}
            of {installWorkflow?.steps?.length} steps completed
          </Text>
        </span>
      </span>

      {/* <div className="">
          {installWorkflow?.status ? (
          <span className="flex gap-2">
          <YAStatus status={installWorkflow?.status?.status} />
          <Text variant="reg-12">
          {sentanceCase(
          installWorkflow?.status?.status_human_description
          ) || 'Waiting on workflow to run.'}
          </Text>
          </span>
          ) : null}
          {installWorkflow?.status?.status === 'error' ? (
          <Text className="ml-9 text-red-800 dark:text-red-500 text-[12px]">
          {sentanceCase(
          installWorkflow?.status?.history?.at(-1)?.status_human_description
          )}
          </Text>
          ) : null}
          </div> */}
    </div>
  )
}
