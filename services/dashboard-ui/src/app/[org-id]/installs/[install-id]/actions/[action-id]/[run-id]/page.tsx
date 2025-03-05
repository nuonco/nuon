import React from 'react'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { CalendarBlank, Timer } from '@phosphor-icons/react/dist/ssr'
import {
  ActionLogsSection,
  ActionWorkflowStatus,
  CancelRunnerJobButton,
  ClickToCopy,
  CodeViewer,
  DashboardContent,
  Duration,
  EventStatus,
  LogStreamProvider,
  Section,
  Text,
  Time,
  ToolTip,
} from '@/components'
import {
  getInstall,
  getAppActionWorkflow,
  getInstallActionWorkflowRun,
} from '@/lib'
import type { TInstallActionWorkflowRun, TActionConfig } from '@/types'
import {
  sentanceCase,
  CANCEL_RUNNER_JOBS,
  humandReadableTriggeredBy,
} from '@/utils'

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

export default withPageAuthRequired(async function InstallWorkflow({ params }) {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const actionWorkflowId = params?.['action-id'] as string
  const actionWorkflowRunId = params?.['run-id'] as string
  const [install, actionWorkflow, workflowRun] = await Promise.all([
    getInstall({ installId, orgId }),
    getAppActionWorkflow({ actionWorkflowId, orgId }),
    getInstallActionWorkflowRun({ installId, orgId, actionWorkflowRunId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}/actions`,
          text: install.name,
        },
        {
          href: `/${orgId}/installs/${install.id}/actions/${actionWorkflowId}`,
          text: `${actionWorkflow?.name}`,
        },
        {
          href: `/${orgId}/installs/${install.id}/actions/${actionWorkflowId}/${workflowRun.id}`,
          text: workflowRun.id,
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
            <Text variant="reg-12">
              {humandReadableTriggeredBy(workflowRun?.triggered_by_type)}
            </Text>
          </span>

          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Install
            </Text>
            <Text variant="med-12">{install.name}</Text>
            <Text variant="mono-12">
              <ToolTip alignment="right" tipContent={install.id}>
                <ClickToCopy>{install.id}</ClickToCopy>
              </ToolTip>
            </Text>
          </span>
          {CANCEL_RUNNER_JOBS &&
          workflowRun?.runner_job?.id &&
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
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="md:col-span-8">
          <LogStreamProvider initLogStream={workflowRun?.log_stream}>
            <ActionLogsSection workflowRun={workflowRun} />
          </LogStreamProvider>
        </div>

        <div className="divide-y flex flex-col md:col-span-4">
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
            <Section className="flex-initial" heading="Workflow outputs">
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
