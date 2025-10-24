'use client'

import classNames from 'classnames'
import { DateTime } from 'luxon'
import { Badge } from '@/components/old/Badge'
import { Link } from '@/components/old/Link'
import { SpinnerSVG } from '@/components/old/Loading'
import { Notice } from '@/components/old/Notice'
import { Time } from '@/components/old/Time'
import { Text } from '@/components/old/Typography'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useQueryParams } from '@/hooks/use-query-params'
import type { TInstallWorkflow } from '@/types'
import { removeSnakeCase, sentanceCase } from '@/utils'
import { InstallWorkflowCancelModal } from './InstallWorkflowCancelModal'
import {
  CaretRightIcon,
  ClockCountdownIcon,
  CheckCircleIcon,
  XCircleIcon,
  WarningIcon,
  WarningDiamondIcon,
  MinusCircleIcon,
  ProhibitIcon,
  RepeatIcon,
  EmptyIcon,
} from '@phosphor-icons/react'

function formatToRelativeDay(dateString: string) {
  const inputDate = DateTime.fromISO(dateString).startOf('day')
  const today = DateTime.now().startOf('day')
  const diffDays = inputDate.diff(today, 'days').days

  if (diffDays === 0) {
    return 'Today'
  } else if (diffDays === -1) {
    return 'Yesterday'
  } else {
    return inputDate.toLocaleString(DateTime.DATE_MED_WITH_WEEKDAY)
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

export interface IInstallWorkflowHistory extends IPollingProps {
  initWorkflows: Array<TInstallWorkflow>
  pagination: {
    offset: string
    limit: number
  }
}

export const InstallWorkflowHistory = ({
  initWorkflows,
  pagination,
  pollInterval = 5000,
  shouldPoll = false,
}: IInstallWorkflowHistory) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const queryParams = useQueryParams({
    offset: pagination?.offset,
    limit: pagination.limit,
  })
  const { data, error } = usePolling({
    dependencies: [queryParams],
    initData: initWorkflows,
    path: `/api/orgs/${org.id}/installs/${install.id}/workflows${queryParams}`,
    pollInterval,
    shouldPoll,
  })

  const workflowHistory = parseInstallWorkflowsByDate(data)

  return (
    <div className="flex flex-col gap-2">
      {error?.error ? (
        <Notice>{error?.error || 'Error loading workflows'}</Notice>
      ) : null}
      {Object.keys(workflowHistory).map((k) => (
        <div key={k} className="flex flex-col gap-2">
          <Text
            variant="med-12"
            className="text-cool-grey-600 dark:text-white/70"
          >
            {formatToRelativeDay(k)}
          </Text>

          <div className="flex flex-col gap-3">
            {workflowHistory[k].map((iw) =>
              iw?.finished || iw?.status?.status === 'cancelled' ? (
                <Link
                  key={iw?.id}
                  className="flex justify-between w-full history-event"
                  href={`/${org?.id}/installs/${iw?.owner_id}/workflows/${iw?.id}`}
                  variant="ghost"
                >
                  <span className="flex gap-4">
                    <YAStatus status={iw.status.status} />
                    <span>
                      <span className="flex gap-2">
                        <Text variant="med-12">
                          {iw?.type === 'action_workflow_run' &&
                          iw?.metadata?.install_action_workflow_name
                            ? sentanceCase(removeSnakeCase(iw?.type)) +
                              ' (' +
                              iw?.metadata?.install_action_workflow_name +
                              ') '
                            : sentanceCase(
                                iw?.name || removeSnakeCase(iw?.type)
                              ) + ' '}
                          {iw?.status?.status}
                        </Text>
                        {iw?.plan_only ? (
                          <Badge
                            className="!text-[10px] p-1 !leading-none ml-2"
                            variant="code"
                            theme="info"
                          >
                            Drift scan
                          </Badge>
                        ) : null}

                        {iw?.type === 'drift_run_reprovision_sandbox' ||
                        iw.type === 'drift_run' ? (
                          <Badge
                            className="!text-[10px] p-1 !leading-none ml-1"
                            variant="code"
                          >
                            cron scheduled
                          </Badge>
                        ) : null}
                      </span>
                      <Text variant="mono-12">{iw?.id}</Text>
                    </span>
                  </span>
                  <Text
                    variant="reg-12"
                    className="text-cool-grey-600 dark:text-white/70 self-end justify-self-end"
                  >
                    <Time
                      time={iw.created_at}
                      format="relative"
                      alignment="right"
                    />
                  </Text>
                </Link>
              ) : (
                <div
                  key={iw?.id}
                  className="flex justify-between w-full history-event p-2"
                >
                  <span className="flex gap-4">
                    <YAStatus status={iw.status.status} />
                    <span>
                      <Link
                        href={`/${org?.id}/installs/${iw?.owner_id}/workflows/${iw?.id}`}
                      >
                        <span className="flex gap-2">
                          <Text variant="med-12">
                            {iw?.type === 'action_workflow_run' &&
                            iw?.metadata?.install_action_workflow_name
                              ? sentanceCase(removeSnakeCase(iw?.type)) +
                                ' (' +
                                iw?.metadata?.install_action_workflow_name +
                                ') '
                              : sentanceCase(
                                  iw?.name || removeSnakeCase(iw?.type)
                                ) + ' '}
                            {iw?.status?.status}
                          </Text>
                          {iw?.plan_only ? (
                            <Badge
                              className="!text-[10px] p-1 !leading-none ml-2"
                              variant="code"
                              theme="info"
                            >
                              Drift scan
                            </Badge>
                          ) : null}

                          {iw?.type === 'drift_run_reprovision_sandbox' ||
                          iw.type === 'drift_run' ? (
                            <Badge
                              className="!text-[10px] p-1 !leading-none ml-1"
                              variant="code"
                            >
                              cron scheduled
                            </Badge>
                          ) : null}
                        </span>
                      </Link>
                      <Text variant="mono-12">{iw?.owner_id}</Text>
                    </span>
                  </span>
                  <div className="flex flex-col gap-0">
                    <div className="flex items-center gap-4">
                      <InstallWorkflowCancelModal
                        buttonClassName="!px-2 !py-0.5"
                        buttonVariant="ghost"
                        installWorkflow={iw}
                      />
                      <Link
                        className="text-sm font-medium"
                        href={`/${org?.id}/installs/${iw?.owner_id}/workflows/${iw?.id}`}
                      >
                        View details <CaretRightIcon />
                      </Link>
                    </div>
                    <Text
                      variant="reg-12"
                      className="text-cool-grey-600 dark:text-white/70 self-end"
                    >
                      <Time time={iw.created_at} format="relative" />
                    </Text>
                  </div>
                </div>
              )
            )}
          </div>
        </div>
      ))}
    </div>
  )
}

export const YAStatus = ({
  status,
  isSkipped = false,
  isRetried = false,
}: {
  status: TInstallWorkflow['status']['status']
  isSkipped?: boolean
  isRetried?: boolean
}) => {
  const isSuccess =
    status === 'active' || status === 'success' || status === 'approved'
  const isError = status === 'error'
  const isProhibit = status === 'outdated'
  const isInProgress = status === 'in-progress'
  const isCanceled = status === 'cancelled'
  const isNotAttempted = status === 'not-attempted'
  const isPendingApproval = status === 'approval-awaiting'
  const isApprovalDenied = status === 'approval-denied'
  const isDiscarded = status === 'discarded'
  const isUserSkipped = status === 'noop'
  const isSystemSkipped = status == 'auto-skipped'
  const isPending =
    !isSuccess && !isError && !isProhibit && !isInProgress && !isSystemSkipped

  const StatusIcon = isSuccess ? (
    <CheckCircleIcon size="18" weight="bold" />
  ) : isError ? (
    <XCircleIcon size="18" weight="bold" />
  ) : isRetried ? (
    <RepeatIcon size="18" weight="bold" />
  ) : isUserSkipped ? (
    <MinusCircleIcon size="18" weight="bold" />
  ) : isProhibit ? (
    <ProhibitIcon size="18" weight="bold" />
  ) : isInProgress ? (
    <SpinnerSVG />
  ) : isCanceled ? (
    <XCircleIcon size="18" weight="bold" />
  ) : isNotAttempted || isDiscarded ? (
    <ProhibitIcon size="18" weight="bold" />
  ) : isApprovalDenied ? (
    <WarningDiamondIcon size="18" weight="bold" />
  ) : isPendingApproval ? (
    <WarningIcon size="18" weight="bold" />
  ) : isSystemSkipped ? (
    <EmptyIcon size="18" weight="bold" />
  ) : (
    <ClockCountdownIcon size="18" weight="bold" />
  )

  return (
    <span
      className={classNames(
        'rounded-full w-[26px] h-[26px] flex items-center justify-center',
        {
          'bg-green-600/15 dark:bg-green-500/15 text-green-800 dark:text-green-500':
            isSuccess && !isSkipped,
          'bg-red-600/15 dark:bg-red-500/15 text-red-800 dark:text-red-500':
            isError,
          'bg-orange-600/15 dark:bg-orange-500/15 text-orange-800 dark:text-orange-500':
            isProhibit || isApprovalDenied,
          'bg-orange-600/15 dark:bg-orange-500/15 text-orange-600 dark:text-orange-300':
            isCanceled || isPendingApproval,
          'bg-blue-600/15 dark:bg-blue-500/15 text-blue-800 dark:text-blue-500':
            isInProgress || isSystemSkipped,
          'bg-cool-grey-600/15 dark:bg-cool-grey-500/15 text-cool-grey-800 dark:text-cool-grey-500':
            isPending || isSkipped || isDiscarded,
        }
      )}
    >
      {isSkipped ? <ProhibitIcon size="18" weight="bold" /> : StatusIcon}
    </span>
  )
}
