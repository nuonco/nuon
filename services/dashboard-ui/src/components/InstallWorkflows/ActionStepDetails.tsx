'use client'

import { useParams } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { CaretRight } from '@phosphor-icons/react'
import { Link } from '@/components/Link'
import { Loading } from '@/components/Loading'
import { Notice } from '@/components/Notice'
import { StatusBadge } from '@/components/Status'
import { EventStatus } from '@/components/Timeline'
import { Duration } from '@/components/Time'
import { Text } from '@/components/Typography'
import type { TActionConfig, TInstallActionWorkflowRun } from '@/types'
import { sentanceCase } from '@/utils'
import type { IPollStepDetails } from './InstallWorkflowSteps'

// hydrate run steps with idx and name
function hydrateRunSteps(
  steps: TInstallActionWorkflowRun['steps'],
  stepConfigs: TActionConfig['steps']
) {
  return steps?.map((step) => {
    const config = stepConfigs?.find((cfg) => cfg.id === step.step_id)
    return {
      name: config?.name,
      idx: config.idx,
      ...step,
    }
  })
}

export const ActionStepDetails: FC<IPollStepDetails> = ({
  step,
  shouldPoll = false,
  pollDuration = 5000,
}) => {
  const params = useParams<Record<'org-id', string>>()
  const orgId = params?.['org-id']
  const [actionRun, setData] = useState<TInstallActionWorkflowRun>()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string>()

  const fetchData = () => {
    fetch(
      `/api/${orgId}/installs/${step?.install_id}/action-workflows/runs/${step?.step_target_id}`
    ).then((r) =>
      r.json().then((res) => {
        setIsLoading(false)
        if (res?.error) {
          setError(res?.error?.error)
        } else {
          setData(res.data)
        }
      })
    )
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
    <>
      {isLoading ? (
        <Loading loadingText="Loading action run details..." variant="page" />
      ) : (
        <>
          {error ? <Notice>{error}</Notice> : null}
          {actionRun ? (
            <div className="flex flex-col border rounded-md shadow">
              <div className="flex items-center justify-between p-3 border-b">
                <Text variant="med-14">Action run</Text>
                <Link
                  className="text-sm gap-0"
                  href={`/${orgId}/installs/${step?.install_id}/actions/${actionRun?.config?.action_workflow_id}/${actionRun?.id}`}
                >
                  View details
                  <CaretRight />
                </Link>
              </div>

              <div className="p-6 flex flex-col gap-4">
                <StatusBadge
                  status={actionRun?.status}
                  description={actionRun?.status_description}
                  label="Action status"
                />
                <div className="flex flex-col gap-2">
                  <Text isMuted className="tracking-wide">
                    Action steps
                  </Text>
                  {hydrateRunSteps(actionRun?.steps, actionRun?.config?.steps)
                    ?.sort(({ idx: a }, { idx: b }) => b - a)
                    ?.reverse()
                    ?.map((actionStep) => {
                      return (
                        <span
                          key={actionStep.id}
                          className="py-2 px-4 border rounded-md flex items-center justify-between"
                        >
                          <span className="flex items-center gap-3">
                            <EventStatus status={actionStep.status} />
                            <Text variant="med-14">{actionStep?.name}</Text>
                          </span>

                          <Text
                            className="flex items-center ml-7"
                            variant="reg-12"
                          >
                            {sentanceCase(actionStep.status)}{' '}
                            {actionStep?.execution_duration > 1000000 ? (
                              <>
                                in{' '}
                                <Duration
                                  nanoseconds={actionStep?.execution_duration}
                                />
                              </>
                            ) : null}
                          </Text>
                        </span>
                      )
                    })}
                </div>
              </div>
            </div>
          ) : null}
        </>
      )}
    </>
  )
}
