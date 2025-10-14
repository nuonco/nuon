import type { Metadata } from 'next'
import { Suspense } from 'react'

import {
  CalendarBlankIcon,
  CaretLeftIcon,
  TimerIcon,
} from '@phosphor-icons/react/dist/ssr'
import { BackLink } from '@/components/common/BackLink'
import { BackToTop } from '@/components/common/BackToTop'
import { PageSection } from '@/components/layout/PageSection'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import {
  getInstallById,
  getInstallSandboxRunById,
  getWorkflowById,
  getOrgById,
} from '@/lib'
import { CANCEL_RUNNER_JOBS, sentanceCase } from '@/utils'
import { Logs, LogsError, LogsSkeleton } from './logs'

// NOTE: old layout stuff
import { ErrorBoundary } from 'react-error-boundary'
import {
  AppSandboxConfig,
  AppSandboxVariables,
  ApprovalStep,
  ClickToCopy,
  DashboardContent,
  Duration,
  InstallWorkflowCancelModal,
  Loading,
  Link,
  LogStreamProvider as OldLogStreamProvider,
  JsonView,
  OperationLogsSection,
  RunnerJobPlanModal,
  SandboxRunStatus,
  Section,
  Text,
  Time,
  ToolTip,
} from '@/components'

export async function generateMetadata({ params }): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['run-id']: runId,
  } = await params
  const [{ data: install }, { data: sandboxRun }] = await Promise.all([
    getInstallById({ installId, orgId }),
    getInstallSandboxRunById({ runId, orgId }),
  ])

  return {
    title: `${sandboxRun.run_type} | ${install.name} | Nuon`,
  }
}

export default async function SandboxRuns({ params }) {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['run-id']: runId,
  } = await params
  const [{ data: install }, { data: sandboxRun }, { data: org }] =
    await Promise.all([
      getInstallById({ installId, orgId }),
      getInstallSandboxRunById({
        orgId,
        runId,
      }),
      getOrgById({ orgId }),
    ])

  const { data: workflow } = await getWorkflowById({
    orgId,
    workflowId: sandboxRun?.install_workflow_id,
  })
  const step = workflow
    ? workflow?.steps
        ?.filter((s) => s?.step_target_id === sandboxRun?.id)
        ?.at(-1)
    : null

  const containerId = 'sandbox-run-page'
  return org?.features?.['stratus-layout'] ? (
    <PageSection className="!p-0" id={containerId} isScrollable>
      {/* old page content */}
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="md:col-span-8">
          <BackLink className="ml-6 mt-6" />

          {workflow &&
          step &&
          step?.approval &&
          !step?.approval?.response &&
          step?.status?.status !== 'auto-skipped' ? (
            <Section
              className="border-b"
              childrenClassName="flex flex-col gap-6"
              heading="Approve change"
            >
              <ApprovalStep
                step={step}
                approval={step.approval}
                workflowId={workflow?.id}
              />
            </Section>
          ) : null}

          <Section>
            <LogStreamProvider
              initLogStream={sandboxRun?.log_stream}
              shouldPoll={sandboxRun?.log_stream?.open}
            >
              <ErrorBoundary fallback={<LogsError />}>
                <Suspense fallback={<LogsSkeleton />}>
                  <Logs
                    logStreamId={sandboxRun?.log_stream?.id}
                    orgId={orgId}
                  />
                </Suspense>
              </ErrorBoundary>
            </LogStreamProvider>
          </Section>

          {workflow &&
          step &&
          step?.approval &&
          step?.approval?.response &&
          step?.status?.status !== 'auto-skipped' ? (
            <Section
              className="border-t"
              childrenClassName="flex flex-col gap-6"
              heading="Approve change"
            >
              <ApprovalStep
                step={step}
                approval={step.approval}
                workflowId={workflow?.id}
              />
            </Section>
          ) : null}
        </div>

        <div className="divide-y flex flex-col md:col-span-4">
          <Section className="flex-initial" heading="Sandbox">
            <div className="flex flex-col gap-3">
              <AppSandboxConfig sandboxConfig={sandboxRun.app_sandbox_config} />
              <AppSandboxVariables
                variables={sandboxRun.app_sandbox_config?.variables}
              />
            </div>
          </Section>

          {sandboxRun?.runner_jobs?.at(0)?.outputs ? (
            <Section className="flex-initial" heading="Sandbox outputs">
              <div className="flex flex-col gap-2">
                <div className="flex items-center justify-between">
                  <Text variant="med-12">Outputs</Text>
                  <ClickToCopy className="hover:bg-black/10 rounded-md p-1 text-sm">
                    <span className="hidden">
                      {JSON.stringify(sandboxRun?.runner_jobs?.at(0).outputs)}
                    </span>
                  </ClickToCopy>
                </div>
                <JsonView data={sandboxRun?.runner_jobs?.at(0)?.outputs} />
              </div>
            </Section>
          ) : null}
        </div>
      </div>

      {/* old page content */}
      <BackToTop containerId={containerId} />
    </PageSection>
  ) : (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}/workflows`,
          text: install.name,
        },
        {
          href: `/${orgId}/installs/${install.id}/sandbox`,
          text: 'Sandbox',
        },
        {
          href: `/${orgId}/installs/${install.id}/${sandboxRun.id}`,
          text: sandboxRun.id,
        },
      ]}
      heading={`${install.name} ${sandboxRun.run_type}`}
      headingUnderline={sandboxRun.id}
      headingMeta={
        sandboxRun?.install_workflow_id ? (
          <Link
            href={`/${orgId}/installs/${installId}/workflows/${sandboxRun?.install_workflow_id}?target=${step?.id}`}
          >
            <CaretLeftIcon />
            View workflow
          </Link>
        ) : null
      }
      meta={
        <div className="flex gap-8 items-center justify-start pb-6">
          <Text>
            <CalendarBlankIcon />
            <Time time={sandboxRun.created_at} />
          </Text>
          {sandboxRun?.runner_jobs?.at(0)?.status === 'finished' ||
          sandboxRun?.runner_jobs?.at(0)?.status === 'failed' ||
          sandboxRun?.runner_jobs?.at(0)?.status === 'cancelled' ? (
            <Text>
              <TimerIcon />
              <Duration
                beginTime={sandboxRun.created_at}
                endTime={sandboxRun.updated_at}
              />
            </Text>
          ) : null}
        </div>
      }
      statues={
        <div className="flex gap-6 items-start justify-start">
          <span className="flex flex-col gap-2">
            <SandboxRunStatus
              descriptionAlignment="right"
              descriptionPosition="bottom"
              initSandboxRun={sandboxRun}
              shouldPoll
            />
          </span>

          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Type
            </Text>
            <Text>{sandboxRun.run_type}</Text>
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
          <ErrorBoundary fallback={<Text>Can&apso;t fetching job plan</Text>}>
            <Suspense
              fallback={
                <Loading variant="stack" loadingText="Loading job plan..." />
              }
            >
              <RunnerJobPlanModal
                runnerJobId={sandboxRun?.runner_jobs?.at(0)?.id}
              />
            </Suspense>
          </ErrorBoundary>
          {CANCEL_RUNNER_JOBS &&
          sandboxRun?.runner_jobs?.at(0)?.status !== 'finished' &&
          sandboxRun?.runner_jobs?.at(0)?.status !== 'failed' &&
          sandboxRun?.runner_jobs?.at(0)?.id &&
          workflow &&
          !workflow?.finished ? (
            <InstallWorkflowCancelModal installWorkflow={workflow} />
          ) : null}
        </div>
      }
    >
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="md:col-span-8">
          {workflow &&
          step &&
          step?.approval &&
          !step?.approval?.response &&
          step?.status?.status !== 'auto-skipped' ? (
            <Section
              className="border-b"
              childrenClassName="flex flex-col gap-6"
              heading="Approve change"
            >
              <ApprovalStep
                step={step}
                approval={step.approval}
                workflowId={workflow?.id}
              />
            </Section>
          ) : null}

          <OldLogStreamProvider initLogStream={sandboxRun?.log_stream}>
            <OperationLogsSection
              heading={sentanceCase(sandboxRun?.run_type) + ' logs'}
            />
          </OldLogStreamProvider>

          {workflow &&
          step &&
          step?.approval &&
          step?.approval?.response &&
          step?.status?.status !== 'auto-skipped' ? (
            <Section
              className="border-t"
              childrenClassName="flex flex-col gap-6"
              heading="Approve change"
            >
              <ApprovalStep
                step={step}
                approval={step.approval}
                workflowId={workflow?.id}
              />
            </Section>
          ) : null}
        </div>

        <div className="divide-y flex flex-col md:col-span-4">
          <Section className="flex-initial" heading="Sandbox">
            <div className="flex flex-col gap-3">
              <AppSandboxConfig sandboxConfig={sandboxRun.app_sandbox_config} />
              <AppSandboxVariables
                variables={sandboxRun.app_sandbox_config?.variables}
              />
            </div>
          </Section>

          {sandboxRun?.runner_jobs?.at(0)?.outputs ? (
            <Section className="flex-initial" heading="Sandbox outputs">
              <div className="flex flex-col gap-2">
                <div className="flex items-center justify-between">
                  <Text variant="med-12">Outputs</Text>
                  <ClickToCopy className="hover:bg-black/10 rounded-md p-1 text-sm">
                    <span className="hidden">
                      {JSON.stringify(sandboxRun?.runner_jobs?.at(0).outputs)}
                    </span>
                  </ClickToCopy>
                </div>
                <JsonView data={sandboxRun?.runner_jobs?.at(0)?.outputs} />
              </div>
            </Section>
          ) : null}
        </div>
      </div>
    </DashboardContent>
  )
}
