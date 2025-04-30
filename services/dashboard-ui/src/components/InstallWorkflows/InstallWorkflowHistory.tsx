'use client'

import classNames from 'classnames'
import { DateTime } from 'luxon'
import { usePathname } from 'next/navigation'
import React, { type FC, useEffect } from 'react'
import {
  ClockCountdown,
  CheckCircle,
  XCircle,
  Prohibit,
} from '@phosphor-icons/react/dist/ssr'
import { revalidateData } from '@/components/actions'
import { Link } from '@/components/Link'
import { SpinnerSVG } from '@/components/Loading'
import { useOrg } from '@/components/Orgs'
import { Time } from '@/components/Time'
import { Text } from '@/components/Typography'
import type { TInstallWorkflow } from '@/types'
import { POLL_DURATION, removeSnakeCase, sentanceCase } from '@/utils'

function formatToRelativeDay(isoDate: string) {
  const inputDate = DateTime.fromISO(isoDate).startOf('day')
  const today = DateTime.now().startOf('day')

  const diffDays = inputDate.diff(today, 'days').days

  if (diffDays === 0) {
    return 'Today'
  } else if (diffDays === -1) {
    return 'Yesterday'
  } else {
    return inputDate.toLocaleString(DateTime.DATETIME_MED_WITH_WEEKDAY)
  }
}

type TInstallWorkflowHistory = Record<string, Array<TInstallWorkflow>>

function parseInstallWorkflowsByDate(
  installWorkflows: Array<TInstallWorkflow>
): TInstallWorkflowHistory {
  return installWorkflows.reduce<TInstallWorkflowHistory>((acc, iw) => {
    const date = iw.created_at.split('T')[0]

    if (!acc[date]) {
      acc[date] = []
    }
    acc[date].push(iw)

    return acc
  }, {})
}

export interface IInstallWorkflowHistory {
  installWorkflows: Array<TInstallWorkflow>
  pollDuration?: number
  shouldPoll?: boolean
}

export const InstallWorkflowHistory: FC<IInstallWorkflowHistory> = ({
  installWorkflows,
  pollDuration = POLL_DURATION,
  shouldPoll = false,
}) => {
  const { org } = useOrg()
  const workflowHistory = parseInstallWorkflowsByDate(installWorkflows)

  const path = usePathname()

  useEffect(() => {
    const refreshData = () => {
      revalidateData({ path })
    }
    if (shouldPoll) {
      const pollBuild = setInterval(refreshData, pollDuration)

      return () => clearInterval(pollBuild)
    }
  }, [installWorkflows, shouldPoll])

  return (
    <div className="flex flex-col gap-2">
      {Object.keys(workflowHistory).map((k) => (
        <div key={k} className="flex flex-col gap-2">
          <Text
            variant="med-12"
            className="text-cool-grey-600 dark:text-white/70"
          >
            {formatToRelativeDay(k)}
          </Text>

          <div>
            {workflowHistory[k].map((iw) => (
              <Link
                key={iw?.id}
                className="flex justify-between w-full history-event"
                href={`/${org?.id}/installs/${iw?.install_id}/history/${iw?.id}`}
                variant="ghost"
              >
                <span className="flex gap-4">
                  <YAStatus status={iw.status.status} />
                  <span>
                    <Text variant="med-12">
                      {sentanceCase(removeSnakeCase(iw?.type))}{' '}
                      {iw?.status?.status}
                    </Text>
                    <Text variant="mono-12">{iw?.install_id}</Text>
                  </span>
                </span>
                <Text
                  variant="reg-12"
                  className="text-cool-grey-600 dark:text-white/70 self-start justify-self-end"
                >
                  <Time time={iw.created_at} format="relative" />
                </Text>
              </Link>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

export const YAStatus: FC<{ status: TInstallWorkflow['status']['status'] }> = ({
  status,
}) => {
  const isSuccess = status === 'active' || status === 'success'
  const isError = status === 'error'
  const isProhibit = status === 'outdated'
  const isInProgress = status === 'in-progress'
  const isCanceled = status === 'cancelled'
  const isNotAttempted = status === 'not-attempted'
  const isPending = !isSuccess && !isError && !isProhibit && !isInProgress

  const StatusIcon = isSuccess ? (
    <CheckCircle size="18" weight="bold" />
  ) : isError ? (
    <XCircle size="18" weight="bold" />
  ) : isProhibit ? (
    <Prohibit size="18" weight="bold" />
  ) : isInProgress ? (
    <SpinnerSVG />
  ) : isCanceled ? (
    <XCircle size="18" weight="bold" />
  ) : isNotAttempted ? (
    <Prohibit size="18" weight="bold" />
  ) : (
    <ClockCountdown size="18" weight="bold" />
  )

  return (
    <span
      className={classNames(
        'rounded-full w-[26px] h-[26px] flex items-center justify-center self-center',
        {
          'bg-green-600/15 dark:bg-green-500/15 text-green-800 dark:text-green-500':
            isSuccess,
          'bg-red-600/15 dark:bg-red-500/15 text-red-800 dark:text-red-500':
            isError,
          'bg-orange-600/15 dark:bg-orange-500/15 text-orange-800 dark:text-orange-500':
            isProhibit,
          'bg-orange-600/15 dark:bg-orange-500/15 text-orange-600 dark:text-orange-300':
            isCanceled,
          'bg-blue-600/15 dark:bg-blue-500/15 text-blue-800 dark:text-blue-500':
            isInProgress,
          'bg-cool-grey-600/15 dark:bg-cool-grey-500/15 text-cool-grey-800 dark:text-cool-grey-500':
            isPending,
        }
      )}
    >
      {StatusIcon}
    </span>
  )
}
