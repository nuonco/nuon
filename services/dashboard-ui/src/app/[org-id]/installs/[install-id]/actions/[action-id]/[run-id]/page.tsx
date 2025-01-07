import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  ClickToCopy,
  DashboardContent,
  Duration,
  LogStreamPoller,
  Section,
  StatusBadge,
  Text,
  Time,
  ToolTip,
  Truncate,
} from '@/components'
import {
  getInstall,
  getAppActionWorkflow,
  getInstallActionWorkflowRun,
  getLogStreamLogs,
  getOrg,
} from '@/lib'
import type { TOTELLog } from '@/types'

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

  const logs = await getLogStreamLogs({
    logStreamId: workflowRun?.log_stream?.id,
    orgId,
  })

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
          <Time time={workflowRun.created_at} />
          <Duration
            beginTime={workflowRun.created_at}
            endTime={workflowRun.updated_at}
          />
        </div>
      }
      statues={
        <div className="flex gap-6 items-start justify-start">
          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Status
            </Text>
            <StatusBadge
              descriptionAlignment="right"
              status={workflowRun?.status}
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
        </div>
      }
    >
      <div className="flex flex-col md:flex-row flex-auto">
        <LogStreamPoller
          heading="Workflow logs"
          initLogStream={workflowRun?.log_stream}
          initLogs={logs as Array<TOTELLog>}
          orgId={orgId}
          logStreamId={workflowRun?.log_stream?.id}
          shouldPoll={Boolean(workflowRun?.log_stream)}
        />

        <div className="divide-y flex flex-col lg:min-w-[450px] lg:max-w-[450px]">
          <Section
            className="flex-initial"
            heading={`${workflowRun?.config?.steps?.length} Steps`}
          >
            <div className="flex flex-col gap-0 divide-y">
              {workflowRun?.config?.steps?.map((step, i) => (
                <span key={step.id} className="py-2">
                  <Text>
                    {i + 1}. {step.name}
                  </Text>
                </span>
              ))}
            </div>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})
