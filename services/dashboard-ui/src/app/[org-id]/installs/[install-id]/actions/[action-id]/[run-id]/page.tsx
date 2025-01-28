// TODO(nnnat): remove no check

import { DateTime } from 'luxon'
import React, { type FC, Suspense } from 'react'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { CalendarBlank, Timer } from '@phosphor-icons/react/dist/ssr'
import {
  ActionWorkflowStatus,
  CancelRunnerJobButton,
  ClickToCopy,
  CodeViewer,
  DashboardContent,
  Duration,
  EventStatus,
  Expand,
  Loading,
  RunnerLogs,
  Section,
  Text,
  Time,
  ToolTip,
  Truncate,
  type TLogRecord,
} from '@/components'
import {
  getInstall,
  getAppActionWorkflow,
  getInstallActionWorkflowRun,
  getLogStreamLogs,
  getOrg,
} from '@/lib'
import type {
  TOTELLog,
  TInstallActionWorkflowRun,
  TActionConfig,
} from '@/types'
import { sentanceCase, CANCEL_RUNNER_JOBS } from '@/utils'

// hydrate run steps with idx and name
function hydrateRunSteps(
  steps: TInstallActionWorkflowRun['steps'],
  stepConfigs: TActionConfig['steps']
) {
  return steps.map((step) => {
    const config = stepConfigs.find((cfg) => cfg.id === step.step_id)
    return {
      name: config?.name,
      idx: config.idx,
      ...step,
    }
  })
}

// convert otel log timestamp from string to milliseconds
function parseOTELLog(logs: Array<TOTELLog>): Array<TLogRecord> {
  return logs?.length
    ? logs?.map((l) => ({
        ...l,
        timestamp: DateTime.fromISO(l.timestamp).toMillis(),
      }))
    : []
}

export default withPageAuthRequired(async function InstallWorkflow({ params }) {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const actionWorkflowId = params?.['action-id'] as string
  const actionWorkflowRunId = params?.['run-id'] as string
  const [org, install, actionWorkflow, workflowRun] = await Promise.all([
    getOrg({ orgId }),
    getInstall({ installId, orgId }),
    getAppActionWorkflow({ actionWorkflowId, orgId }),
    getInstallActionWorkflowRun({ installId, orgId, actionWorkflowRunId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/installs`, text: 'Installs' },
        {
          href: `/${org.id}/installs/${install.id}/actions`,
          text: install.name,
        },
        {
          href: `/${org.id}/installs/${install.id}/actions/${actionWorkflowId}`,
          text: `${actionWorkflow?.name}`,
        },
      ]}
      heading={`${actionWorkflow?.name} execution`}
      headingUnderline={actionWorkflowId}
      meta={
        <div className="flex gap-8 items-center justify-start pb-6">
          <Text>
            <CalendarBlank size={14} />
            <Time time={workflowRun.created_at} />
          </Text>
          <Text>
            <Timer size={14} />
            <Duration nanoseconds={workflowRun.execution_time} />
          </Text>
        </div>
      }
      statues={
        <div className="flex gap-6 items-start justify-start">
          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Status
            </Text>
            <ActionWorkflowStatus
              descriptionAlignment="right"
              actionWorkflowRun={workflowRun}
              shouldPoll
            />
          </span>

          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Trigger type
            </Text>
            <Text variant="mono-12">{workflowRun.trigger_type}</Text>
          </span>

          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Install
            </Text>
            <Text variant="med-12">{install.name}</Text>
            <Text variant="mono-12">
              <ToolTip alignment="right" tipContent={install.id}>
                <ClickToCopy>
                  <Truncate variant="small">{install.id}</Truncate>
                </ClickToCopy>
              </ToolTip>
            </Text>
          </span>
          {CANCEL_RUNNER_JOBS &&
          (workflowRun?.status === 'queued' ||
            workflowRun?.status === 'in-progress') ? (
            <CancelRunnerJobButton
              jobType="sandbox-run"
              runnerJobId={workflowRun?.runner_job?.id}
              orgId={orgId}
            />
          ) : null}
        </div>
      }
    >
      <div className="flex flex-col md:flex-row flex-auto">
        <Section heading="Workflow step logs" className="border-r">
          <Suspense
            fallback={<Loading loadingText="Loading action workflow steps" />}
          >
            <LoadLogs orgId={orgId} workflowRun={workflowRun} />
          </Suspense>
        </Section>
        <div className="divide-y flex flex-col lg:min-w-[450px] lg:max-w-[450px]">
          <Section
            className="flex-initial"
            heading={`${workflowRun?.steps?.reduce((count, step) => {
              if (step.status === 'finished' || step.status === 'error') count++
              return count
            }, 0)} of ${workflowRun?.config?.steps?.length} Steps`}
          >
            <div className="flex flex-col gap-2">
              {hydrateRunSteps(workflowRun?.steps, workflowRun?.config?.steps)
                ?.sort(({ idx: a }, { idx: b }) => b - a)
                ?.reverse()
                ?.map((step) => {
                  return (
                    <span key={step.id} className="py-2">
                      <span className="flex items-center gap-3">
                        <EventStatus status={step.status} />
                        <Text variant="med-14">{step?.name}</Text>
                      </span>

                      <Text className="flex items-center ml-7" variant="reg-12">
                        {sentanceCase(step.status)}{' '}
                        {step?.execution_duration > 1000 ? (
                          <>
                            in{' '}
                            <Duration nanoseconds={step?.execution_duration} />
                          </>
                        ) : (
                          'did not attempt'
                        )}
                      </Text>
                    </span>
                  )
                })}
            </div>
          </Section>
          {workflowRun?.runner_job?.outputs ? (
            <Section heading="Workflow outputs">
              <CodeViewer
                initCodeSource={JSON.stringify(
                  workflowRun?.runner_job?.outputs || {},
                  null,
                  2
                )}
                language="json"
              />
            </Section>
          ) : null}
        </div>
      </div>
    </DashboardContent>
  )
})

const LoadLogs: FC<{
  orgId: string
  workflowRun: TInstallActionWorkflowRun
}> = async ({ orgId, workflowRun }) => {
  const logs = await getLogStreamLogs({
    logStreamId: workflowRun?.log_stream?.id,
    orgId,
  }).catch(console.error)

  if (!logs) {
    return <Text variant="reg-14">Waiting on action workflow to run.</Text>
  }

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

  if (Object.keys(logSteps).length === 0) {
    return <Text variant="reg-14">Waiting on action workflow logs.</Text>
  }

  return (
    <div className="flex flex-col gap-3">
      {Object.keys(logSteps).map((step) => {
        const workflowStep = workflowRun?.steps?.find(
          (s) => s?.id === logSteps[step]?.at(0)?.log_attributes?.step_run_id
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
              <RunnerLogs
                heading={step}
                logs={parseOTELLog(logSteps[step])}
                withOutBorder
              />
            }
          />
        )
      })}
    </div>
  )
}
