import type { Metadata } from 'next'
import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { CalendarBlankIcon, TimerIcon } from '@phosphor-icons/react/dist/ssr'
import {
  DashboardContent,
  Duration,
  ErrorFallback,
  Loading,
  LogStreamProvider,
  OperationLogsSection,
  Section,
  Time,
  Text,
  ToolTip,
  Truncate,
} from '@/components'
import { BuildDetails } from '@/components/Components/BuildDetails'
import { getAppById, getComponentBuildById } from '@/lib'
import { ComponentConfig } from './config'

export async function generateMetadata({ params }): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['component-id']: componentId,
    ['build-id']: buildId,
  } = await params
  const { data: build } = await getComponentBuildById({
    componentId,
    buildId,
    orgId,
  })

  return {
    title: `Build | ${build?.component_name} | Nuon`,
  }
}

export default async function AppComponent({ params }) {
  const {
    ['org-id']: orgId,
    ['app-id']: appId,
    ['component-id']: componentId,
    ['build-id']: buildId,
  } = await params

  const [{ data: app }, { data: build }] = await Promise.all([
    getAppById({ appId, orgId }),
    getComponentBuildById({ componentId, buildId, orgId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/apps`, text: 'Apps' },
        { href: `/${orgId}/apps/${app?.id}`, text: app?.name },
        { href: `/${orgId}/apps/${app?.id}/components`, text: 'Components' },
        {
          href: `/${orgId}/apps/${app?.id}/components/${build.component_id}`,
          text: build?.component_name,
        },
        {
          href: `/${orgId}/apps/${app?.id}/components/${build.component_id}/builds/${build.id}`,
          text: 'Build',
        },
      ]}
      heading={`${build?.component_name} build`}
      headingUnderline={build.id}
      meta={
        <div className="flex gap-8 items-center justify-start pb-6">
          <Text>
            <CalendarBlankIcon />
            <Time time={build.created_at} />
          </Text>
          <Text>
            <TimerIcon />
            <Duration beginTime={build.created_at} endTime={build.updated_at} />
          </Text>
        </div>
      }
      statues={<BuildDetails initBuild={build} shouldPoll />}
    >
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="md:col-span-8">
          {build?.log_stream ? (
            <LogStreamProvider initLogStream={build?.log_stream}>
              <OperationLogsSection heading="Build logs" />
            </LogStreamProvider>
          ) : (
            <Section heading="Build logs">
              <Text>Waiting on log stream</Text>
            </Section>
          )}
        </div>
        <div className="divide-y flex flex-col md:col-span-4">
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
                    {build.vcs_connection_commit?.message?.length >= 32 ? (
                      <ToolTip
                        tipContent={build.vcs_connection_commit?.message}
                        alignment="right"
                        position="top"
                      >
                        <Truncate variant="small">
                          {build.vcs_connection_commit?.message}
                        </Truncate>
                      </ToolTip>
                    ) : (
                      build?.vcs_connection_commit?.message
                    )}
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
                <ComponentConfig
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
}
