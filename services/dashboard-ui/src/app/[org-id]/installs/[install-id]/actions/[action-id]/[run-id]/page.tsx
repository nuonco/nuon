import type { Metadata } from 'next'
import { Suspense } from 'react'
import {
  CalendarBlankIcon,
  CaretLeftIcon,
  TimerIcon,
} from '@phosphor-icons/react/dist/ssr'
import { BackLink } from '@/components/common/BackLink'
import { BackToTop } from '@/components/common/BackToTop'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { PageSection } from '@/components/layout/PageSection'
import { Text } from '@/components/common/Text'
import {
  getInstallActionById,
  getInstallActionRunById,
  getInstallById,
  getWorkflowById,
  getOrgById,
} from '@/lib'
import { InstallActionRunProvider } from '@/providers/install-action-run-provider'
import { CANCEL_RUNNER_JOBS } from '@/utils'

// NOTE: old layout stuff
import { ErrorBoundary } from 'react-error-boundary'
import {
  ActionTriggerType,
  ActionLogsSection,
  ActionWorkflowStatus,
  ClickToCopy,
  DashboardContent,
  Duration,
  Link as OldLink,
  Loading,
  LogStreamProvider,
  RunnerJobPlanModal,
  Text as OldText,
  Time,
  ToolTip,
} from '@/components'
import { InstallActionCancelButton } from '@/components/InstallActionRunCancelButton'
import { InstallActionRunDetails } from '@/components/InstallActionRunDetails'

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

export default async function InstallActionRunPage({ params }) {
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
    { data: org },
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
    getOrgById({ orgId }),
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

  const containerId = 'action-run-page'
  return (
    <InstallActionRunProvider
      initInstallActionRun={installActionRun}
      shouldPoll
    >
      {org?.features?.['stratus-layout'] ? (
        <PageSection className="!p-0 !gap-0" id={containerId} isScrollable>
          {/* old page layout */}
          <div className="p-6 border-b flex justify-between">
            <HeadingGroup>
              <BackLink className="mb-6" />
              <Text variant="base" weight="strong">
                {installAction.action_workflow?.name}
              </Text>
              <ID>{actionId}</ID>
              <div className="flex gap-8 items-center justify-start mt-2">
                <Text className="!flex items-center gap-1">
                  <CalendarBlankIcon size={14} />
                  <Time time={installActionRun.created_at} />
                </Text>
                <Text className="!flex items-center gap-1">
                  <TimerIcon size={14} />
                  <Duration nanoseconds={installActionRun.execution_time} />
                </Text>
              </div>
              {installActionRun?.install_workflow_id ? (
                <Link
                  className="text-xs mt-2"
                  href={`/${orgId}/installs/${installId}/workflows/${installActionRun?.install_workflow_id}?target=${step?.id}`}
                >
                  <CaretLeftIcon />
                  View workflow
                </Link>
              ) : null}
            </HeadingGroup>
            <div className="flex gap-6 items-start justify-start">
              <span className="flex flex-col gap-2">
                <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                  Status
                </OldText>
                <ActionWorkflowStatus descriptionAlignment="right" />
              </span>
              <span className="flex flex-col gap-2">
                <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                  Trigger type
                </OldText>
                <ActionTriggerType
                  triggerType={installActionRun?.triggered_by_type}
                  componentName={installActionRun?.run_env_vars?.COMPONENT_NAME}
                  componentPath={`/${orgId}/installs/${installId}/components/${installActionRun?.run_env_vars?.COMPONENT_ID}`}
                />
              </span>

              <span className="flex flex-col gap-2">
                <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                  Install
                </OldText>
                <OldText variant="med-12">{install.name}</OldText>
                <OldText variant="mono-12">
                  <ToolTip alignment="right" tipContent={install.id}>
                    <ClickToCopy>{install.id}</ClickToCopy>
                  </ToolTip>
                </OldText>
              </span>
              {installActionRun?.runner_job?.id ? (
                <ErrorBoundary
                  fallback={<OldText>Can&apso;t fetching job plan</OldText>}
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
          </div>
          <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
            <div className="md:col-span-8">
              <LogStreamProvider initLogStream={installActionRun?.log_stream}>
                <ActionLogsSection />
              </LogStreamProvider>
            </div>
            <InstallActionRunDetails />
          </div>
          {/* old page layout */}
          <BackToTop containerId={containerId} />
        </PageSection>
      ) : (
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
              <OldLink
                href={`/${orgId}/installs/${installId}/workflows/${installActionRun?.install_workflow_id}?target=${step?.id}`}
              >
                <CaretLeftIcon />
                View workflow
              </OldLink>
            ) : null
          }
          meta={
            <div className="flex gap-8 items-center justify-start pb-6">
              <OldText>
                <CalendarBlankIcon size={14} />
                <Time time={installActionRun.created_at} />
              </OldText>
              <OldText>
                <TimerIcon size={14} />
                <Duration nanoseconds={installActionRun.execution_time} />
              </OldText>
            </div>
          }
          statues={
            <div className="flex gap-6 items-start justify-start">
              <span className="flex flex-col gap-2">
                <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                  Status
                </OldText>
                <ActionWorkflowStatus descriptionAlignment="right" />
              </span>
              <span className="flex flex-col gap-2">
                <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                  Trigger type
                </OldText>
                <ActionTriggerType
                  triggerType={installActionRun?.triggered_by_type}
                  componentName={installActionRun?.run_env_vars?.COMPONENT_NAME}
                  componentPath={`/${orgId}/installs/${installId}/components/${installActionRun?.run_env_vars?.COMPONENT_ID}`}
                />
              </span>

              <span className="flex flex-col gap-2">
                <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                  Install
                </OldText>
                <OldText variant="med-12">{install.name}</OldText>
                <OldText variant="mono-12">
                  <ToolTip alignment="right" tipContent={install.id}>
                    <ClickToCopy>{install.id}</ClickToCopy>
                  </ToolTip>
                </OldText>
              </span>
              {installActionRun?.runner_job?.id ? (
                <ErrorBoundary
                  fallback={<OldText>Can&apso;t fetching job plan</OldText>}
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
      )}
    </InstallActionRunProvider>
  )
}
