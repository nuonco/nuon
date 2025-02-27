import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { CalendarBlank, Timer } from '@phosphor-icons/react/dist/ssr'
import {
  BuildStatus,
  CancelRunnerJobButton,
  ClickToCopy,
  ComponentConfiguration,
  DashboardContent,
  Duration,
  ErrorFallback,
  LogStreamPoller,
  Loading,
  Section,
  Time,
  Text,
  ToolTip,
  Truncate,
} from '@/components'
import {
  getApp,
  getComponent,
  getComponentBuild,
  getComponentConfig,
  getLogStreamLogs,
} from '@/lib'
import type { TBuild, TOTELLog } from '@/types'
import { CANCEL_RUNNER_JOBS } from '@/utils'

export default withPageAuthRequired(async function AppComponent({ params }) {
  const appId = params?.['app-id'] as string
  const buildId = params?.['build-id'] as string
  const componentId = params?.['component-id'] as string
  const orgId = params?.['org-id'] as string

  const build = await getComponentBuild({ buildId, orgId })
  const [app, component, logs] = await Promise.all([
    getApp({ appId, orgId }),
    getComponent({ componentId, orgId }),
    getLogStreamLogs({ orgId, logStreamId: build.log_stream?.id }).catch(
      console.error
    ),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/apps`, text: 'Apps' },
        { href: `/${orgId}/apps/${app.id}/components`, text: app.name },
        {
          href: `/${orgId}/apps/${app.id}/components/${build.component_id}`,
          text: component.name,
        },
        {
          href: `/${orgId}/apps/${app.id}/components/${build.component_id}/builds/${build.id}`,
          text: `${component.name} build`,
        },
      ]}
      heading={`${component.name} build`}
      headingUnderline={build.id}
      meta={
        <div className="flex gap-8 items-center justify-start pb-6">
          <Text>
            <CalendarBlank />
            <Time time={build.created_at} />
          </Text>
          <Text>
            <Timer />
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
                <ClickToCopy>{build.component_id}</ClickToCopy>
              </ToolTip>
            </Text>
          </span>
          {CANCEL_RUNNER_JOBS &&
          build?.status !== 'active' &&
          build?.status !== 'error' ? (
            <CancelRunnerJobButton
              jobType="build"
              runnerJobId={build?.runner_job?.id}
              orgId={orgId}
            />
          ) : null}
        </div>
      }
    >
      <div className="flex flex-col lg:flex-row flex-auto">
        <LogStreamPoller
          heading="Build logs"
          initLogs={logs as Array<TOTELLog>}
          initLogStream={build.log_stream}
          orgId={orgId}
          logStreamId={build.log_stream?.id}
          shouldPoll={Boolean(build?.log_stream)}
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
                  <Text>
                    <ToolTip
                      tipContent={build.vcs_connection_commit?.message}
                      alignment="right"
                      position="top"
                    >
                      <Truncate variant="extra-large">
                        {build.vcs_connection_commit?.message}
                      </Truncate>
                    </ToolTip>
                  </Text>
                </span>
              </div>
            </Section>
          )}

          <Section className="flex-initial" heading="Component config">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading
                    variant="stack"
                    loadingText="Loading component config..."
                  />
                }
              >
                <LoadComponentConfig
                  componentId={componentId}
                  componentConfigId={build?.component_config_connection_id}
                  orgId={orgId}
                />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})

const LoadBuildLogs: FC<{ build: TBuild; orgId: string }> = async ({
  build,
  orgId,
}) => {
  const logs = await getLogStreamLogs({
    orgId,
    logStreamId: build.log_stream?.id,
  }).catch(console.error)
  return (
    <LogStreamPoller
      heading="Build logs"
      initLogs={logs as Array<TOTELLog>}
      initLogStream={build.log_stream}
      orgId={orgId}
      logStreamId={build.log_stream?.id}
      shouldPoll={Boolean(build?.log_stream)}
    />
  )
}

const LoadComponentConfig: FC<{
  componentId: string
  componentConfigId: string
  orgId: string
}> = async ({ componentId, componentConfigId, orgId }) => {
  const componentConfig = await getComponentConfig({
    componentId,
    componentConfigId,
    orgId,
  })
  return <ComponentConfiguration config={componentConfig} />
}
