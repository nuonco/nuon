'use client'

import { usePathname } from 'next/navigation'
import React, { type FC, useEffect } from 'react'
import { useUser } from '@auth0/nextjs-auth0'
import { ArrowSquareOut } from '@phosphor-icons/react'
import { revalidateData } from '@/components/actions'
import { Link } from '@/components/Link'
import { Text } from '@/components/Typography'
import type { TInstallWorkflow } from '@/types'
import { SHORT_POLL_DURATION } from '@/utils'

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
  const { user, isLoading } = useUser()

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
    <div className="">
      <span className="flex w-full justify-between flex-wrap">
        <span className="flex flex-col gap-0">
          <span className="flex items-center gap-4 ml-auto">
            <span>&#x1F680;</span>
            <progress
              className="rounded-lg [&::-webkit-progress-bar]:rounded-lg [&::-webkit-progress-value]:rounded-lg   [&::-webkit-progress-bar]:bg-cool-grey-300 [&::-webkit-progress-value]:bg-green-400 [&::-moz-progress-bar]:bg-green-400 [&::-webkit-progress-value]:transition-all [&::-webkit-progress-value]:duration-500 [&::-moz-progress-bar]:transition-all [&::-moz-progress-bar]:duration-500 h-[8px]"
              max={installWorkflow?.steps?.length}
              value={
                installWorkflow?.steps?.filter(
                  (s) =>
                    s?.status?.status === 'success' ||
                    s?.status?.status === 'active' ||
                    s?.status?.status === 'error' ||
                    s?.status?.status === 'approved'
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
                  s?.status?.status === 'error' ||
                  s?.status?.status === 'approved'
              ).length
            }{' '}
            of {installWorkflow?.steps?.length} steps completed{' '}
            {installWorkflow?.steps?.filter(
              (s) => s?.status?.status === 'discarded'
            ).length ? (
              <>
                ,{' '}
                {
                  installWorkflow?.steps?.filter(
                    (s) => s?.status?.status === 'discarded'
                  ).length
                }{' '}
                steps discarded
              </>
            ) : null}
          </Text>
        </span>
      </span>

      {!isLoading && user?.email?.endsWith('@nuon.co') ? (
        <Link
          className="text-base gap-2 mt-3 ml-auto"
          href={`/admin/temporal/namespaces/installs/workflows/${installWorkflow?.owner_id}-execute-workflow-${installWorkflow?.id}`}
          target="_blank"
        >
          View in Temporal <ArrowSquareOut />
        </Link>
      ) : null}
    </div>
  )
}
