import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { FiCloud, FiClock } from 'react-icons/fi'
import {
  AppSandboxConfig,
  AppSandboxVariables,
  CancelRunnerJobButton,
  ClickToCopy,
  CodeViewer,
  DashboardContent,
  Duration,
  LogStreamPoller,
  SandboxRunStatus,
  Section,
  Text,
  Truncate,
  Time,
  ToolTip,
} from '@/components'
import {
  getInstall,
  getLogStreamLogs,
  getInstallSandboxRun,
  getOrg,
} from '@/lib'
import type { TOTELLog } from '@/types'
import { CANCEL_RUNNER_JOBS } from '@/utils'

export default withPageAuthRequired(async function SandboxRuns({ params }) {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const runId = params?.['run-id'] as string
  const sandboxRun = await getInstallSandboxRun({
    installId,
    orgId,
    installSandboxRunId: runId,
  })
  const [install, org, logs] = await Promise.all([
    getInstall({ installId, orgId }),
    getOrg({ orgId }),
    getLogStreamLogs({
      orgId,
      logStreamId: sandboxRun?.log_stream?.id,
    }).catch(console.error),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/installs`, text: 'Installs' },
        {
          href: `/${org.id}/installs/${install.id}/history`,
          text: install.name,
        },
        {
          href: `/${org.id}/installs/${install.id}/runs/${sandboxRun.id}`,
          text: `${install.name} ${sandboxRun.run_type}`,
        },
      ]}
      heading={`${install.name} ${sandboxRun.run_type}`}
      headingUnderline={sandboxRun.id}
      meta={
        <div className="flex gap-8 items-center justify-start pb-6">
          <Text>
            <FiCloud />
            <Time time={sandboxRun.created_at} />
          </Text>
          <Text>
            <FiClock />
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
                <ClickToCopy>
                  <Truncate variant="small">{install.id}</Truncate>
                </ClickToCopy>
              </ToolTip>
            </Text>
          </span>
          {CANCEL_RUNNER_JOBS &&
          sandboxRun?.status !== 'active' &&
          sandboxRun?.status !== 'error' &&
          sandboxRun?.runner_job?.id ? (
            <CancelRunnerJobButton
              jobType="sandbox-run"
              runnerJobId={sandboxRun?.runner_job?.id}
              orgId={orgId}
            />
          ) : null}
        </div>
      }
    >
      <div className="flex flex-col lg:flex-row flex-auto">
        <LogStreamPoller
          heading={sandboxRun?.run_type + ' logs'}
          initLogStream={sandboxRun?.log_stream}
          initLogs={logs as Array<TOTELLog>}
          orgId={orgId}
          logStreamId={sandboxRun?.log_stream?.id}
          shouldPoll={Boolean(sandboxRun?.log_stream)}
        />

        <div
          className="divide-y flex flex-col lg:min-w-[450px]
lg:max-w-[450px]"
        >
          <Section className="flex-initial" heading="Sandbox">
            <AppSandboxConfig sandboxConfig={sandboxRun.app_sandbox_config} />
            <AppSandboxVariables
              variables={sandboxRun.app_sandbox_config?.variables}
            />
          </Section>

          {sandboxRun?.runner_job?.outputs ? (
            <Section heading="Sandbox outputs">
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
        </div>
      </div>
    </DashboardContent>
  )
})
