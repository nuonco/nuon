import type { Metadata } from 'next'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { CalendarBlank, CaretLeft, Timer } from '@phosphor-icons/react/dist/ssr'
import {
  AppSandboxConfig,
  AppSandboxVariables,
  CancelRunnerJobButton,
  ClickToCopy,
  CodeViewer,
  DashboardContent,
  Duration,
  InstallDeployIntermediateData,
  InstallWorkflowCancelModal,
  Link,
  LogStreamProvider,
  OperationLogsSection,
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
import { CANCEL_RUNNER_JOBS, sentanceCase } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const runId = params?.['run-id'] as string
  const [install, sandboxRun] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallSandboxRun({ installId, installSandboxRunId: runId, orgId }),
  ])

  return {
    title: `${install.name} | ${sandboxRun.run_type}`,
  }
}

export default withPageAuthRequired(async function SandboxRuns({ params }) {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const runId = params?.['run-id'] as string
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

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}/history`,
          text: install.name,
        },
        {
          href: `/${orgId}/installs/${install.id}/sandbox`,
          text: 'Sandbox',
        },
        {
          href: `/${orgId}/installs/${install.id}/runs/${sandboxRun.id}`,
          text: sandboxRun.id,
        },
      ]}
      heading={`${install.name} ${sandboxRun.run_type}`}
      headingUnderline={sandboxRun.id}
      headingMeta={
        sandboxRun?.install_workflow_id ? (
          <Link
            href={`/${orgId}/installs/${installId}/history/${sandboxRun?.install_workflow_id}?target=${runId}`}
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
          <Text>
            <Timer />
            <Duration
              beginTime={sandboxRun.created_at}
              endTime={sandboxRun.updated_at}
            />
          </Text>
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
          {CANCEL_RUNNER_JOBS &&
          sandboxRun?.runner_job?.status !== 'finished' &&
          sandboxRun?.runner_job?.status !== 'failed' &&
          sandboxRun?.runner_job?.id &&
          installWorkflow &&
          !installWorkflow?.finished ? (
            <InstallWorkflowCancelModal installWorkflow={installWorkflow} />
          ) : null}
        </div>
      }
    >
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="md:col-span-8">
          <LogStreamProvider initLogStream={sandboxRun?.log_stream}>
            <OperationLogsSection
              heading={sentanceCase(sandboxRun?.run_type) + ' logs'}
            />
          </LogStreamProvider>
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

          {sandboxRun?.runner_job?.outputs ? (
            <Section className="flex-initial" heading="Sandbox outputs">
              <div className="flex flex-col gap-2">
                <div className="flex items-center justify-between">
                  <Text variant="med-12">Outputs</Text>
                  <ClickToCopy className="hover:bg-black/10 rounded-md p-1 text-sm">
                    <span className="hidden">
                      {JSON.stringify(sandboxRun?.runner_job.outputs)}
                    </span>
                  </ClickToCopy>
                </div>
                <CodeViewer
                  initCodeSource={JSON.stringify(
                    sandboxRun?.runner_job?.outputs,
                    null,
                    2
                  )}
                  language="json"
                />
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
              runnerJobId={sandboxRun?.runner_job?.id}
              />
              </Suspense>
              </ErrorBoundary> */}
        </div>
      </div>
    </DashboardContent>
  )
})

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
