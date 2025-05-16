'use client'

import { useParams } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { CaretRight } from '@phosphor-icons/react'
import { Link } from '@/components/Link'
import { Loading } from '@/components/Loading'
import { Notice } from '@/components/Notice'
import { StatusBadge } from '@/components/Status'
import { Time } from '@/components/Time'
import { Text } from '@/components/Typography'
import type { TRunner, TRunnerHeartbeat } from '@/types'
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
        setRunner(run.data)
      }

      if (heart?.error) {
        setError(heart?.error?.error)
      } else {
        setRunnerHeartbeat(heart.data)
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
        <Loading loadingText="Loading runner details" variant="page" />
      ) : (
        <>
          {error ? <Notice>{error}</Notice> : null}{' '}
          {runner ? (
            <div className="flex flex-col border rounded-md shadow">
              <div className="flex items-center justify-between p-3 border-b">
                <Text variant="med-14">Install runner</Text>
                <Link
                  className="text-sm gap-0"
                  href={`/${orgId}/installs/${step?.install_id}/runner`}
                >
                  View details
                  <CaretRight />
                </Link>
              </div>
              <div className="p-6 grid grid-cols-4">
                {runnerHeartbeat ? (
                  <>
                    <StatusBadge
                      description={runner?.status_description}
                      status={runner?.status}
                      label="Status"
                    />
                    <span className="flex flex-col gap-2">
                      <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                        Started at
                      </Text>
                      <Text>
                        <Time
                          time={runnerHeartbeat?.started_at}
                          format="default"
                          variant="med-12"
                        />
                      </Text>
                    </span>
                    <span className="flex flex-col gap-2">
                      <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                        Version
                      </Text>
                      <Text variant="med-12">{runnerHeartbeat?.version}</Text>
                    </span>
                  </>
                ) : null}
                <span className="flex flex-col gap-2">
                  <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                    Platform
                  </Text>
                  <Text variant="med-12">{platform}</Text>
                </span>
              </div>
            </div>
          ) : null}
        </>
      )}
    </div>
  )
}
