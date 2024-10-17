import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { FiCloud, FiClock } from 'react-icons/fi'
import {
  DashboardContent,
  Duration,
  ComponentConfiguration,
  Heading,
  RunnerLogs,
  StatusBadge,
  Time,
  Text,
  ToolTip,
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

export default withPageAuthRequired(
  async function AppComponent({ params }) {
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
          { href: `/${org.id}/apps/${app.id}`, text: app.name },
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
            <Text variant="caption">
              <FiCloud />
              <Time time={build.created_at} variant="caption" />
            </Text>
            <Text variant="caption">
              <FiClock />
              <Duration
                beginTime={build.created_at}
                endTime={build.updated_at}
                variant="caption"
              />
            </Text>
          </div>
        }
        statues={
          <div className="flex gap-6 items-start justify-start">
            <span className="flex flex-col gap-2">
              <Text variant="overline">Status</Text>
              <StatusBadge
                descriptionAlignment="right"
                descriptionPosition="bottom"
                description={build.status_description}
                status={build.status}
              />
            </span>

            <span className="flex flex-col gap-2">
              <Text variant="overline">Component</Text>
              <Text variant="label">{component.name}</Text>
              <Text variant="id">{build.component_id}</Text>
            </span>
          </div>
        }
      >
        <div className="flex flex-col lg:flex-row flex-auto">
          <RunnerLogs heading="Build logs" logs={logs as Array<TOTELLog>} />

          <div
            className="divide-y flex flex-col lg:min-w-[450px]
lg:max-w-[450px]"
          >
            {build.vcs_connection_commit && (
              <section className="flex flex-col gap-6 px-6 py-8">
                <Heading>Commit details</Heading>

                <div className="flex gap-6 items-start justify-start">
                  <span className="flex flex-col gap-2">
                    <Text variant="overline">SHA</Text>
                    <ToolTip tipContent={build.vcs_connection_commit?.sha}>
                      <Text
                        className="truncate text-ellipsis w-16"
                        variant="id"
                      >
                        {build.vcs_connection_commit?.sha}
                      </Text>
                    </ToolTip>
                  </span>

                  {build.vcs_connection_commit?.author_name !== '' && (
                    <span className="flex flex-col gap-2">
                      <Text variant="overline">Author</Text>
                      <Text variant="caption">
                        {build.vcs_connection_commit?.author_name}
                      </Text>
                    </span>
                  )}

                  <span className="flex flex-col gap-2">
                    <Text variant="overline">Message</Text>
                    <Text variant="caption">
                      {build.vcs_connection_commit?.message}
                    </Text>
                  </span>
                </div>
              </section>
            )}
            <section className="flex flex-col gap-6 px-6 py-8">
              <Heading>Component configuration</Heading>

              <ComponentConfiguration config={componentConfig} />
            </section>
          </div>
        </div>
      </DashboardContent>
    )
  },
  { returnTo: '/' }
)
