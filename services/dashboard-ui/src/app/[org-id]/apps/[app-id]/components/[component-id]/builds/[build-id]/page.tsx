import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { FiCloud, FiClock } from 'react-icons/fi'
import {
  BuildStatus,
  ClickToCopy,
  DashboardContent,
  Duration,
  ComponentConfiguration,
  Heading,
  RunnerLogsPoller,
  Section,
  Time,
  Text,
  ToolTip,
  Truncate,
} from '@/components'
import {
  getApp,
  getBuild,
  getComponent,
  getComponentConfig,
  getRunnerLogs,
  getOrg,
} from '@/lib'
import type { TOTELLog } from '@/types'

export default withPageAuthRequired(async function AppComponent({ params }) {
  const appId = params?.['app-id'] as string
  const buildId = params?.['build-id'] as string
  const componentId = params?.['component-id'] as string
  const orgId = params?.['org-id'] as string

  const build = await getBuild({ buildId, orgId })
  const [app, component, componentConfig, org, logs] = await Promise.all([
    getApp({ appId, orgId }),
    getComponent({ componentId, orgId }),
    getComponentConfig({
      componentId,
      componentConfigId: build.component_config_connection_id,
      orgId,
    }),
    getOrg({ orgId }),
    getRunnerLogs({
      jobId: build.runner_job?.id,
      runnerId: build.runner_job?.runner_id,
      orgId,
    }).catch(console.error),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/apps`, text: 'Apps' },
        { href: `/${org.id}/apps/${app.id}/components`, text: app.name },
        {
          href: `/${org.id}/apps/${app.id}/components/${build.component_id}`,
          text: component.name,
        },
        {
          href: `/${org.id}/apps/${app.id}/components/${build.component_id}/builds/${build.id}`,
          text: `${component.name} build`,
        },
      ]}
      heading={`${component.name} build`}
      headingUnderline={build.id}
      meta={
        <div className="flex gap-8 items-center justify-start pb-6">
          <Text>
            <FiCloud />
            <Time time={build.created_at} />
          </Text>
          <Text>
            <FiClock />
            <Duration beginTime={build.created_at} endTime={build.updated_at} />
          </Text>
        </div>
      }
      statues={
        <div className="flex gap-6 items-start justify-start">
          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Status
            </Text>
            <BuildStatus
              descriptionAlignment="right"
              initBuild={build}
              shouldPoll
            />
          </span>

          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Component
            </Text>
            <Text variant="med-12">{component.name}</Text>
            <Text variant="mono-12">
              <ToolTip alignment="right" tipContent={build.component_id}>
                <ClickToCopy>
                  <Truncate variant="small">{build.component_id}</Truncate>
                </ClickToCopy>
              </ToolTip>
            </Text>
          </span>
        </div>
      }
    >
      <div className="flex flex-col lg:flex-row flex-auto">
        <RunnerLogsPoller
          heading="Build logs"
          initJob={build?.runner_job}
          initLogs={logs as Array<TOTELLog>}
          jobId={build?.runner_job?.id}
          orgId={orgId}
          runnerId={build?.runner_job?.runner_id}
          shouldPoll={Boolean(build?.runner_job)}
        />

        <div
          className="divide-y flex flex-col lg:min-w-[450px]
lg:max-w-[450px]"
        >
          {build.vcs_connection_commit && (
            <Section className="flex-initial" heading="Commit details">
              <div className="flex gap-6 items-start justify-start">
                <span className="flex flex-col gap-2">
                  <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                    SHA
                  </Text>
                  <ToolTip tipContent={build.vcs_connection_commit?.sha}>
                    <Text
                      className="truncate text-ellipsis w-16"
                      variant="mono-12"
                    >
                      {build.vcs_connection_commit?.sha}
                    </Text>
                  </ToolTip>
                </span>

                {build.vcs_connection_commit?.author_name !== '' && (
                  <span className="flex flex-col gap-2">
                    <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                      Author
                    </Text>
                    <Text>{build.vcs_connection_commit?.author_name}</Text>
                  </span>
                )}

                <span className="flex flex-col gap-2">
                  <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                    Message
                  </Text>
                  <Text>{build.vcs_connection_commit?.message}</Text>
                </span>
              </div>
            </Section>
          )}

          <Section heading="Component config">
            <ComponentConfiguration config={componentConfig} />
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})
