import type { Metadata } from 'next'
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { CalendarBlank, CaretLeft, Timer } from '@phosphor-icons/react/dist/ssr'
import {
  AppSandboxConfig,
  AppSandboxVariables,
  ApprovalStep,
  ClickToCopy,
  CodeViewer,
  DashboardContent,
  Duration,
  ErrorFallback,
  InstallDeployIntermediateData,
  InstallWorkflowCancelModal,
  Loading,
  Link,
  LogStreamProvider,
  JsonView,
  OperationLogsSection,
  RunnerJobPlanModal,
  SandboxRunStatus,
  Section,
  Text,
  Time,
  ToolTip,
} from '@/components'
import {
  getInstall,
  getInstallSandboxRun,
  getInstallWorkflow,
  getRunnerJobPlan,
} from '@/lib'
import { CANCEL_RUNNER_JOBS, sentanceCase, nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['run-id']: runId,
  } = await params
  const [install, sandboxRun] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallSandboxRun({ installId, installSandboxRunId: runId, orgId }),
  ])

  return {
    title: `${install.name} | ${sandboxRun.run_type}`,
  }
}

export default async function SandboxRuns({ params }) {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['run-id']: runId,
  } = await params
  const [install, sandboxRun] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallSandboxRun({
      installId,
      orgId,
      installSandboxRunId: runId,
    }),
  ])

  const installWorkflow = await getInstallWorkflow({
    orgId,
    installWorkflowId: sandboxRun?.install_workflow_id,
  }).catch(console.error)
  const step = installWorkflow
    ? installWorkflow?.steps?.filter((s) => s?.step_target_id === sandboxRun?.id)?.at(-1)
    : null

  return (
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
            <CaretLeft />
            View workflow
          </Link>
        ) : null
      }
      meta={
        <div className="flex gap-8 items-center justify-start pb-6">
          <Text>
            <CalendarBlank />
            <Time time={sandboxRun.created_at} />
          </Text>
          {sandboxRun?.runner_jobs?.at(0)?.status === 'finished' ||
          sandboxRun?.runner_jobs?.at(0)?.status === 'failed' ||
          sandboxRun?.runner_jobs?.at(0)?.status === 'cancelled' ? (
            <Text>
              <Timer />
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
          installWorkflow &&
          !installWorkflow?.finished ? (
            <InstallWorkflowCancelModal installWorkflow={installWorkflow} />
          ) : null}
        </div>
      }
    >
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="md:col-span-8">
          {installWorkflow &&
          step &&
          step?.approval &&
          !step?.approval?.response &&
          step?.status?.status !== 'auto-skipped'? (
            <Section
              className="border-b"
              childrenClassName="flex flex-col gap-6"
              heading="Approve change"
            >
              <ApprovalStep
                step={step}
                approval={step.approval}
                workflowId={installWorkflow?.id}
              />
            </Section>
          ) : null}

          <LogStreamProvider initLogStream={sandboxRun?.log_stream}>
            <OperationLogsSection
              heading={sentanceCase(sandboxRun?.run_type) + ' logs'}
            />
          </LogStreamProvider>

          {installWorkflow &&
          step &&
          step?.approval &&
          step?.approval?.response &&
          step?.status?.status !== 'auto-skipped'? (
            <Section
              className="border-t"
              childrenClassName="flex flex-col gap-6"
              heading="Approve change"
            >
              <ApprovalStep
                step={step}
                approval={step.approval}
                workflowId={installWorkflow?.id}
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

          {/* <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
              fallback={
              <Section>
              <Loading
              loadingText="Loading intermediate data..."
              variant="stack"
              />
              </Section>
              }
              >
              <LoadSandboxRunPlan
              install={install}
              orgId={orgId}
              runnerJobId={sandboxRun?.runner_jobs?.at(0)?.id}
              />
              </Suspense>
              </ErrorBoundary> */}
        </div>
      </div>
    </DashboardContent>
  )
}

const LoadSandboxRunPlan = async ({ install, orgId, runnerJobId }) => {
  const plan = await getRunnerJobPlan({ orgId, runnerJobId }).catch(
    console.error
  )
  return plan ? (
    <Section heading="Sandbox indermediate data">
      {JSON.stringify(plan)}
      <InstallDeployIntermediateData
        install={install}
        data={plan?.waypointPlan?.variables?.intermediaData}
      />
    </Section>
  ) : null
}
