'use client'

import React, { type FC, useEffect, useState } from 'react'
import { Button } from '@/components/Button'
import { Section } from '@/components/Card'
import { Expand } from '@/components/Expand'
import { RunnerLogs, parseOTELLog } from '@/components/RunnerLogs'
import { Duration } from '@/components/Time'
import { EventStatus } from '@/components/Timeline'
import { Text } from '@/components/Typography'
import type { TOTELLog, TLogStream, TInstallActionWorkflowRun } from '@/types'
import { LOG_POLL_DURATION, SHORT_POLL_DURATION } from '@/utils'

export interface InstallIActionLogStreamPoller {
  initLogStream: TLogStream
  initLogs: Array<TOTELLog>
  orgId: string
  logStreamId: string
  workflowRun: TInstallActionWorkflowRun
  shouldPoll?: boolean
}

export const InstallActionLogStreamPoller: FC<
  InstallIActionLogStreamPoller
> = ({ initLogStream, shouldPoll, ...props }) => {
  const [logStream, updateLogStream] = useState(initLogStream)
  const [isPolling, setPolling] = useState(shouldPoll)

  useEffect(() => {
    const fetchLogStream = () => {
      fetch(`/api/${props.orgId}/log-streams/${props.logStreamId}`)
        .then((res) => res.json().then((l) => updateLogStream(l)))
        .catch(console.error)
    }

    if (shouldPoll) {
      const pollJob = setInterval(fetchLogStream, SHORT_POLL_DURATION)

      if (!logStream.open) {
        setPolling(false)
        clearInterval(pollJob)
      } else {
        setPolling(true)
      }

      return () => {
        clearInterval(pollJob)
      }
    }
  }, [logStream, shouldPoll])

  return <InstallActionLogPoller {...props} shouldPoll={isPolling} />
}

const InstallActionLogPoller: FC<
  Omit<InstallIActionLogStreamPoller, 'initLogStream'>
> = ({ initLogs, orgId, logStreamId, workflowRun, shouldPoll = false }) => {
  const [showStream, setShowStream] = useState(false)
  const [logs, updateLogs] = useState(parseOTELLog(initLogs))

  useEffect(() => {
    const fetchLogs = () => {
      fetch(`/api/${orgId}/log-streams/${logStreamId}/logs`)
        .then((res) => res.json().then((l) => updateLogs(parseOTELLog(l))))
        .catch(console.error)
    }

    if (shouldPoll) {
      const pollLogs = setInterval(fetchLogs, LOG_POLL_DURATION)

      return () => clearInterval(pollLogs)
    }
  }, [logs, shouldPoll])

  const logSteps = (logs as unknown as Array<TOTELLog>).reduce((acc, log) => {
    if (log.log_attributes?.workflow_step_name) {
      if (acc?.[log.log_attributes?.workflow_step_name]) {
        acc[log.log_attributes?.workflow_step_name].push(log)
      } else {
        acc = { ...acc, [log.log_attributes?.workflow_step_name]: [] }
        acc[log.log_attributes?.workflow_step_name].push(log)
      }
    }

    return acc
  }, {})

  const ToggleButton = (
    <Button
      className="text-sm"
      onClick={() => {
        setShowStream(!showStream)
      }}
    >
      {showStream ? 'Display step logs' : 'Display full log stream'}
    </Button>
  )

  return !logs ? (
    <Section
      heading="Workflow step logs"
      className="border-r"
      actions={ToggleButton}
    >
      <Text variant="reg-14">Waiting on action workflow to run.</Text>
    </Section>
  ) : showStream ? (
    <RunnerLogs
      actions={ToggleButton}
      heading="Workflow step logs"
      logs={logs}
    />
  ) : (
    <Section
      heading="Workflow step logs"
      className="border-r"
      actions={ToggleButton}
    >
      {Object.keys(logSteps).length === 0 ? (
        <Text variant="reg-14">Waiting on action workflow logs.</Text>
      ) : (
        <div className="flex flex-col gap-3">
          {Object.keys(logSteps).map((step) => {
            const workflowStep = workflowRun?.steps?.find(
              (s) =>
                s?.id === logSteps[step]?.at(0)?.log_attributes?.step_run_id
            )

            return (
              <Expand
                parentClass="border rounded divide-y"
                headerClass="px-3 py-2"
                id={step}
                key={step}
                heading={
                  <span className="flex gap-3 items-center">
                    <EventStatus status={workflowStep?.status} />
                    <Text variant="med-14">{step}</Text>
                    {workflowStep?.status === 'finished' ||
                    workflowStep.status === 'error' ? (
                      <Duration
                        className="ml-2"
                        nanoseconds={workflowStep?.execution_duration}
                        variant="reg-12"
                      />
                    ) : null}
                  </span>
                }
                isOpen
                expandContent={
                  <div className="max-h-[500px] overflow-x-hidden overflow-y-auto">
                    <RunnerLogs
                      heading={step}
                      logs={parseOTELLog(logSteps[step])}
                      withOutBorder
                    />
                  </div>
                }
              />
            )
          })}
        </div>
      )}
    </Section>
  )
}
