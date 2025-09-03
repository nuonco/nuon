import type { Metadata } from 'next'
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { CalendarBlank, CaretLeft, Timer } from '@phosphor-icons/react/dist/ssr'
import {
  ActionTriggerType,
  ActionLogsSection,
  ActionWorkflowStatus,
  Badge,
  ClickToCopy,
  DashboardContent,
  Duration,
  EventStatus,
  InstallWorkflowCancelModal,
  JsonView,
  Link,
  Loading,
  LogStreamProvider,
  RunnerJobPlanModal,
  Section,
  Text,
  Time,
  ToolTip,
} from '@/components'
import {
  getInstall,
  getAppActionWorkflow,
  getInstallActionWorkflowRun,
  getInstallWorkflow,
} from '@/lib'
import type { TInstallActionWorkflowRun, TActionConfig } from '@/types'
import { sentanceCase, CANCEL_RUNNER_JOBS, nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['action-id']: actionWorkflowId,
    ['run-id']: actionWorkflowRunId,
  } = await params
  const [action, run] = await Promise.all([
    getAppActionWorkflow({ actionWorkflowId, orgId }),
    getInstallActionWorkflowRun({ actionWorkflowRunId, installId, orgId }),
  ])

  return {
    title: `${action.name} | ${run.trigger_type} run`,
  }
}

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

export default async function InstallWorkflow({ params }) {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['action-id']: actionWorkflowId,
    ['run-id']: actionWorkflowRunId,
  } = await params
  const [install, actionWorkflow, workflowRun] = await Promise.all([
    getInstall({ installId, orgId }),
    getAppActionWorkflow({ actionWorkflowId, orgId }),
    getInstallActionWorkflowRun({ installId, orgId, actionWorkflowRunId }),
  ])

  const installWorkflow = await getInstallWorkflow({
    orgId,
    installWorkflowId: workflowRun?.install_workflow_id,
  }).catch(console.error)
  const step = installWorkflow
    ? installWorkflow?.steps
        ?.filter((s) => s?.step_target_id === workflowRun?.id)
        ?.at(-1)
    : null

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}`,
          text: install.name,
        },
        {
          href: `/${orgId}/installs/${install.id}/actions`,
          text: 'Actions',
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
      headingMeta={
        workflowRun?.install_workflow_id ? (
          <Link
            href={`/${orgId}/installs/${installId}/workflows/${workflowRun?.install_workflow_id}?target=${step?.id}`}
          >
            <CaretLeft />
            View workflow
          </Link>
        ) : null
      }
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
            <ActionTriggerType
              triggerType={workflowRun?.triggered_by_type}
              componentName={workflowRun?.run_env_vars?.COMPONENT_NAME}
              componentPath={`/${orgId}/installs/${installId}/components/${workflowRun?.run_env_vars?.COMPONENT_ID}`}
            />
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
          {workflowRun?.runner_job?.id ? (
            <ErrorBoundary fallback={<Text>Can&apso;t fetching job plan</Text>}>
              <Suspense
                fallback={
                  <Loading variant="stack" loadingText="Loading job plan..." />
                }
              >
                <RunnerJobPlanModal runnerJobId={workflowRun?.runner_job?.id} />
              </Suspense>
            </ErrorBoundary>
          ) : null}
          {CANCEL_RUNNER_JOBS &&
          workflowRun?.runner_job?.id &&
          (workflowRun?.status === 'queued' ||
            workflowRun?.status === 'in-progress') &&
          installWorkflow &&
          !installWorkflow?.finished ? (
            <InstallWorkflowCancelModal installWorkflow={installWorkflow} />
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
                        {step?.execution_duration > 1000000 ? (
                          <>
                            in{' '}
                            <Duration nanoseconds={step?.execution_duration} />
                          </>
                        ) : null}
                      </Text>
                    </span>
                  )
                })}
            </div>
          </Section>
          {workflowRun?.runner_job?.outputs ? (
            <Section className="flex-initial" heading="Workflow outputs">
              <JsonView data={workflowRun?.runner_job?.outputs} />
            </Section>
          ) : null}
        </div>
      </div>
    </DashboardContent>
  )
}
