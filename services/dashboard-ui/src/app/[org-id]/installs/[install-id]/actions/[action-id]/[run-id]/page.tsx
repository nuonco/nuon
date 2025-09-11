import type { Metadata } from 'next'
import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { CalendarBlank, CaretLeft, Timer } from '@phosphor-icons/react/dist/ssr'
import {
  ActionTriggerType,
  ActionLogsSection,
  ActionWorkflowStatus,
  ClickToCopy,
  DashboardContent,
  Duration,
  Link,
  Loading,
  LogStreamProvider,
  RunnerJobPlanModal,
  Text,
  Time,
  ToolTip,
} from '@/components'
import { InstallActionCancelButton } from '@/components/InstallActionRunCancelButton'
import { InstallActionRunDetails } from '@/components/InstallActionRunDetails'
import {
  getInstallActionById,
  getInstallActionRunById,
  getInstallById,
  getWorkflowById,
} from '@/lib'
import { InstallActionRunProvider } from '@/providers/install-action-run-provider'
import { CANCEL_RUNNER_JOBS } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['action-id']: actionId,
    ['run-id']: runId,
  } = await params
  const [{ data: installActionRun }, { data: installAction }] =
    await Promise.all([
      getInstallActionRunById({
        installId,
        orgId,
        runId,
      }),
      getInstallActionById({
        actionId,
        installId,
        orgId,
      }),
    ])

  return {
    title: `${installAction?.action_workflow?.name} | ${installActionRun.trigger_type} run`,
  }
}

export default async function InstallWorkflow({ params }) {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['action-id']: actionId,
    ['run-id']: runId,
  } = await params
  const [
    { data: install },
    { data: installActionRun },
    { data: installAction },
  ] = await Promise.all([
    getInstallById({ installId, orgId }),
    getInstallActionRunById({
      installId,
      orgId,
      runId,
    }),
    getInstallActionById({
      actionId,
      installId,
      orgId,
    }),
  ])

  const { data: installWorkflow } = await getWorkflowById({
    orgId,
    workflowId: installActionRun?.install_workflow_id,
  })
  const step = installWorkflow
    ? installWorkflow?.steps
        ?.filter((s) => s?.step_target_id === installActionRun?.id)
        ?.at(-1)
    : null

  return (
    <InstallActionRunProvider
      initInstallActionRun={installActionRun}
      shouldPoll
    >
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
            href: `/${orgId}/installs/${install.id}/actions/${actionId}`,
            text: `${installAction?.action_workflow?.name}`,
          },
          {
            href: `/${orgId}/installs/${install.id}/actions/${actionId}/${installActionRun.id}`,
            text: installActionRun.id,
          },
        ]}
        heading={`${installAction?.action_workflow?.name} execution`}
        headingUnderline={actionId}
        headingMeta={
          installActionRun?.install_workflow_id ? (
            <Link
              href={`/${orgId}/installs/${installId}/workflows/${installActionRun?.install_workflow_id}?target=${step?.id}`}
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
              <Time time={installActionRun.created_at} />
            </Text>
            <Text>
              <Timer size={14} />
              <Duration nanoseconds={installActionRun.execution_time} />
            </Text>
          </div>
        }
        statues={
          <div className="flex gap-6 items-start justify-start">
            <span className="flex flex-col gap-2">
              <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                Status
              </Text>
              <ActionWorkflowStatus descriptionAlignment="right" />
            </span>
            <span className="flex flex-col gap-2">
              <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                Trigger type
              </Text>
              <ActionTriggerType
                triggerType={installActionRun?.triggered_by_type}
                componentName={installActionRun?.run_env_vars?.COMPONENT_NAME}
                componentPath={`/${orgId}/installs/${installId}/components/${installActionRun?.run_env_vars?.COMPONENT_ID}`}
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
            {installActionRun?.runner_job?.id ? (
              <ErrorBoundary
                fallback={<Text>Can&apso;t fetching job plan</Text>}
              >
                <Suspense
                  fallback={
                    <Loading
                      variant="stack"
                      loadingText="Loading job plan..."
                    />
                  }
                >
                  <RunnerJobPlanModal
                    runnerJobId={installActionRun?.runner_job?.id}
                  />
                </Suspense>
              </ErrorBoundary>
            ) : null}
            {CANCEL_RUNNER_JOBS ? (
              <InstallActionCancelButton workflow={installWorkflow} />
            ) : null}
          </div>
        }
      >
        <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
          <div className="md:col-span-8">
            <LogStreamProvider initLogStream={installActionRun?.log_stream}>
              <ActionLogsSection />
            </LogStreamProvider>
          </div>
          <InstallActionRunDetails />
        </div>
      </DashboardContent>
    </InstallActionRunProvider>
  )
}
