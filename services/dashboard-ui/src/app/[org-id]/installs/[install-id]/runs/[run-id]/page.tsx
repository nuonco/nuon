import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { FiCloud, FiClock } from 'react-icons/fi'
import {
  AppSandboxConfig,
  AppSandboxVariables,
  ClickToCopy,
  DashboardContent,
  Duration,
  RunnerLogsPoller,
  SandboxRunStatus,
  Section,
  Text,
  Truncate,
  Time,
  ToolTip,
} from '@/components'
import { getInstall, getRunnerLogs, getSandboxRun, getOrg } from '@/lib'
import type { TOTELLog } from '@/types'

export default withPageAuthRequired(async function SandboxRuns({ params }) {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const runId = params?.['run-id'] as string
  const sandboxRun = await getSandboxRun({ installId, orgId, runId })
  const [install, org, logs] = await Promise.all([
    getInstall({ installId, orgId }),
    getOrg({ orgId }),
    getRunnerLogs({
      jobId: sandboxRun?.runner_job?.id,
      orgId,
      runnerId: sandboxRun?.runner_job?.runner_id,
    }).catch(console.error),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/installs`, text: 'Installs' },
        {
          href: `/${org.id}/installs/${install.id}`,
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
        </div>
      }
    >
      <div className="flex flex-col lg:flex-row flex-auto">
        <RunnerLogsPoller
          heading={sandboxRun?.run_type + ' logs'}
          initJob={sandboxRun?.runner_job}
          initLogs={logs as Array<TOTELLog>}
          jobId={sandboxRun?.runner_job?.id}
          orgId={orgId}
          runnerId={sandboxRun?.runner_job?.runner_id}
          shouldPoll={Boolean(sandboxRun?.runner_job)}
        />

        <div
          className="divide-y flex flex-col lg:min-w-[450px]
lg:max-w-[450px]"
        >
          <Section heading="Sandbox">
            <AppSandboxConfig sandboxConfig={sandboxRun.app_sandbox_config} />
            <AppSandboxVariables
              variables={sandboxRun.app_sandbox_config?.variables}
            />
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})
