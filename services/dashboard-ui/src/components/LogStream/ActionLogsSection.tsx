'use client'

import React, { type FC, useState } from 'react'
import { LogsControls } from './LogsControls'
import { LogsPreview } from './LogsPreview'
import { LogsModal } from './LogsModal'
import { OperationLogsSection } from './OperationLogsSection'
import { useLogs } from './logs-context'
import { LogsViewerProvider } from './logs-viewer-context'
import { Button } from '@/components/Button'
import { Section } from '@/components/Card'
import { ErrorFallback } from '@/components/ErrorFallback'
import { Expand } from '@/components/Expand'
import { Loading } from '@/components/Loading'
import { Duration } from '@/components/Time'
import { EventStatus } from '@/components/Timeline'
import { Text } from '@/components/Typography'
import { useInstallActionRun } from '@/hooks/use-install-action-run'
import type { TOTELLog } from '@/types'

export const ActionLogsSection = () => {
  const { installActionRun } = useInstallActionRun()
  const { error, isLoading, logs } = useLogs()
  const [showStream, setShowStream] = useState(false)

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
      {showStream ? 'Display step logs' : 'Display log stream'}
    </Button>
  )

  return showStream ? (
    <OperationLogsSection
      heading="Workflow log stream"
      actions={ToggleButton}
    />
  ) : (
    <Section heading="Workflow step logs" actions={ToggleButton}>
      {error ? (
        <div>
          <ErrorFallback error={error} resetErrorBoundary={() => {}} />
        </div>
      ) : isLoading && !logs?.length ? (
        <div className="mt-12">
          <Loading loadingText="Loading workflow logs..." variant="stack" />
        </div>
      ) : Object.keys(logSteps).length === 0 ? (
        <Text variant="reg-14">Waiting on action workflow step logs.</Text>
      ) : (
        <div className="flex flex-col gap-3">
          {Object.keys(logSteps).map((step) => {
            const actionRunStep = installActionRun?.steps?.find(
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
                    <EventStatus status={actionRunStep?.status} />
                    <Text variant="med-14">{step}</Text>
                    {actionRunStep?.status === 'finished' ||
                    actionRunStep.status === 'error' ? (
                      <Duration
                        className="ml-2"
                        nanoseconds={actionRunStep?.execution_duration}
                        variant="reg-12"
                      />
                    ) : null}
                  </span>
                }
                isOpen
                expandContent={
                  <LogsViewerProvider>
                    <div className="max-h-[500px] overflow-x-hidden overflow-y-auto flex flex-col gap-3 p-3">
                      <div className="flex items-center justify-end divide-x">
                        <LogsControls />
                        <div className="ml-4 pl-4">
                          <LogsModal
                            heading={`${step} logs`}
                            logs={logSteps[step]}
                          />
                        </div>
                      </div>
                      <LogsPreview logs={logSteps[step]} />
                    </div>
                  </LogsViewerProvider>
                }
              />
            )
          })}
        </div>
      )}
    </Section>
  )
}
