'use client'

import { CaretRightIcon } from '@phosphor-icons/react'
import { Link } from '@/components/old/Link'
import { Loading } from '@/components/old/Loading'
import { Notice } from '@/components/old/Notice'
import { StatusBadge } from '@/components/old/Status'
import { Time } from '@/components/old/Time'
import { ID, Text } from '@/components/old/Typography'
import { useOrg } from '@/hooks/use-org'
import { usePolling } from '@/hooks/use-polling'
import type { TRunner, TRunnerMngHeartbeat, TRunnerSettings } from '@/types'
import { isLessThan15SecondsOld } from '@/utils/time-utils'
import type { IPollStepDetails } from './InstallWorkflowSteps'

interface IRunnerStepDetails extends IPollStepDetails {
  platform?: string
}

export const RunnerStepDetails = ({
  step,
  shouldPoll = false,
  platform = 'aws',
}: IRunnerStepDetails) => {
  const { org } = useOrg()
  const {
    data: runner,
    error,
    isLoading: isRunnerLoading,
  } = usePolling<TRunner>({
    initIsLoading: true,
    path: `/api/orgs/${org.id}/runners/${step?.step_target_id}`,
    pollInterval: 20000,
    shouldPoll,
  })

  const { data: settings, isLoading: isSettingsLoading } =
    usePolling<TRunnerSettings>({
      initIsLoading: true,
      path: `/api/orgs/${org.id}/runners/${step?.step_target_id}/settings`,
      pollInterval: 20000,
      shouldPoll,
    })

  const { data: heartbeats, isLoading: isHeartbeatLoading } =
    usePolling<TRunnerMngHeartbeat>({
      path: `/api/orgs/${org.id}/runners/${step?.step_target_id}/heartbeat`,
      pollInterval: 5000,
      shouldPoll,
    })

  const runnerHeartbeat =
    heartbeats?.install ?? heartbeats?.org ?? heartbeats?.[''] ?? undefined

  return (
    <div className="flex flex-col gap-8">
      {(isRunnerLoading && !runner) || (isSettingsLoading && !settings) ? (
        <div className="border rounded-md p-6">
          <Loading loadingText="Loading runner details..." variant="stack" />
        </div>
      ) : (
        <>
          {error?.error ? <Notice>{error?.error}</Notice> : null}{' '}
          {runner !== null ? (
            <div className="flex flex-col border rounded-md shadow">
              <div className="flex items-center justify-between p-3 border-b">
                <Text variant="med-14">Install runner</Text>
                <Link
                  className="text-sm gap-0"
                  href={`/${org.id}/installs/${step?.owner_id}/runner`}
                >
                  View details
                  <CaretRightIcon />
                </Link>
              </div>

              <>
                <div className="flex gap-8 items-start justify-start flex-wrap p-6">
                  <span className="flex flex-col gap-2">
                    <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                      Status
                    </Text>
                    <StatusBadge
                      status={
                        runner?.status === 'active' ? 'healthy' : 'unhealthy'
                      }
                      description={runner?.status_description}
                      descriptionAlignment="left"
                      isWithoutBorder
                    />

                    {runnerHeartbeat &&
                    isLessThan15SecondsOld(runnerHeartbeat?.created_at) ? (
                      <StatusBadge
                        status="connected"
                        description="Connected to runner"
                        descriptionAlignment="left"
                        isWithoutBorder
                      />
                    ) : (
                      <StatusBadge
                        status="not-connected"
                        description="Waiting on connection"
                        descriptionAlignment="left"
                        isWithoutBorder
                      />
                    )}
                  </span>
                  {runnerHeartbeat ? (
                    <>
                      <span className="flex flex-col gap-2">
                        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                          Started at
                        </Text>
                        <Text>
                          <Time
                            time={runnerHeartbeat?.started_at}
                            format="default"
                          />
                        </Text>
                      </span>
                      <span className="flex flex-col gap-2">
                        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                          Version
                        </Text>
                        <Text>{runnerHeartbeat?.version}</Text>
                      </span>
                    </>
                  ) : null}
                  <span className="flex flex-col gap-2">
                    <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                      Platform
                    </Text>
                    <Text>{settings?.platform || 'aws'}</Text>
                  </span>
                  <span className="flex flex-col gap-2">
                    <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                      ID
                    </Text>
                    <ID id={runner?.id} />
                  </span>
                </div>
              </>
            </div>
          ) : null}
        </>
      )}
    </div>
  )
}
