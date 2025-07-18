'use client'

import { useParams } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { CaretRight } from '@phosphor-icons/react'
import { Link } from '@/components/Link'
import { Loading } from '@/components/Loading'
import { Notice } from '@/components/Notice'
import { StatusBadge } from '@/components/Status'
import { Time } from '@/components/Time'
import { ID, Text } from '@/components/Typography'
import type { TRunner, TRunnerHeartbeat } from '@/types'
import { isLessThan15SecondsOld } from '@/utils/time-utils'
import type { IPollStepDetails } from './InstallWorkflowSteps'

interface IRunnerStepDetails extends IPollStepDetails {
  platform?: string
}

export const RunnerStepDetails: FC<IRunnerStepDetails> = ({
  step,
  shouldPoll = false,
  pollDuration = 5000,
  platform = 'aws',
}) => {
  const params = useParams<Record<'org-id', string>>()
  const orgId = params?.['org-id']
  const [runner, setRunner] = useState<TRunner>()
  const [runnerHeartbeat, setRunnerHeartbeat] = useState<TRunnerHeartbeat>()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string>()

  const fetchData = () => {
    Promise.all([
      fetch(`/api/${orgId}/runners/${step?.step_target_id}`).then((r) =>
        r.json()
      ),

      fetch(
        `/api/${orgId}/runners/${step?.step_target_id}/latest-heart-beat`
      ).then((r) => r.json()),
    ]).then(([run, heart]) => {
      setIsLoading(false)
      if (run?.error) {
        setError(run?.error?.error)
      } else {
        setError(undefined)
        setRunner(run.data)
      }

      if (heart?.error) {
        if (heart?.status !== 404) {
          setError(heart?.error?.error)
        }
      } else {
        setError(undefined)
        setRunnerHeartbeat(heart?.data)
      }
    })
  }

  useEffect(() => {
    fetchData()
  }, [])

  useEffect(() => {
    if (shouldPoll) {
      const pollData = setInterval(fetchData, pollDuration)

      return () => clearInterval(pollData)
    }
  }, [shouldPoll])

  return (
    <div className="flex flex-col gap-8">
      {isLoading ? (
        <div className="border rounded-md p-6">
          <Loading loadingText="Loading runner details..." variant="stack" />
        </div>
      ) : (
        <>
          {error ? <Notice>{error}</Notice> : null}{' '}
          {runner ? (
            <div className="flex flex-col border rounded-md shadow">
              <div className="flex items-center justify-between p-3 border-b">
                <Text variant="med-14">Install runner</Text>
                <Link
                  className="text-sm gap-0"
                  href={`/${orgId}/installs/${step?.owner_id}/runner`}
                >
                  View details
                  <CaretRight />
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
                      shouldPoll
                      isWithoutBorder
                    />

                    {runnerHeartbeat &&
                    isLessThan15SecondsOld(runnerHeartbeat?.created_at) ? (
                      <StatusBadge
                        status="connected"
                        description="Connected to runner"
                        descriptionAlignment="left"
                        shouldPoll
                        isWithoutBorder
                      />
                    ) : (
                      <StatusBadge
                        status="not-connected"
                        description="Waiting on connection"
                        descriptionAlignment="left"
                        shouldPoll
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
                    <Text>{platform}</Text>
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
