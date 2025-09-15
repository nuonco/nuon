import type { Metadata } from 'next'
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { CalendarBlankIcon, TimerIcon } from '@phosphor-icons/react/dist/ssr'
import {
  BuildStatus,
  CancelRunnerJobButton,
  ClickToCopy,
  ComponentConfiguration,
  DashboardContent,
  Duration,
  ErrorFallback,
  Loading,
  LogStreamProvider,
  OperationLogsSection,
  RunnerJobPlanModal,
  Section,
  Time,
  Text,
  ToolTip,
  Truncate,
} from '@/components'
import { getAppById, getComponentById, getComponentBuildById } from '@/lib'
import type { TComponentConfig } from '@/types'
import { CANCEL_RUNNER_JOBS, nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['component-id']: componentId } = await params
  const { data: component} = await getComponentById({ componentId, orgId })

  return {
    title: `Build | ${component.name} | Nuon`,
  }
}

export default async function AppComponent({ params }) {
  const {
    ['org-id']: orgId,
    ['app-id']: appId,
    ['component-id']: componentId,
    ['build-id']: buildId,
  } = await params

  const [{ data: app}, {data: build}, {data: component}] = await Promise.all([
    getAppById({ appId, orgId }),
    getComponentBuildById({ componentId, buildId, orgId }),
    getComponentById({ componentId, orgId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/apps`, text: 'Apps' },
        { href: `/${orgId}/apps/${app.id}`, text: app.name },
        { href: `/${orgId}/apps/${app.id}/components`, text: 'Components' },
        {
          href: `/${orgId}/apps/${app.id}/components/${build.component_id}`,
          text: component.name,
        },
        {
          href: `/${orgId}/apps/${app.id}/components/${build.component_id}/builds/${build.id}`,
          text: 'Build',
        },
      ]}
      heading={`${component.name} build`}
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

          <ErrorBoundary fallback={<Text>Can&apso;t fetching job plan</Text>}>
            <Suspense
              fallback={
                <Loading variant="stack" loadingText="Loading job plan..." />
              }
            >
              <RunnerJobPlanModal
                runnerJobId={build?.runner_job?.id}
              />
            </Suspense>
          </ErrorBoundary>

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
                        <Truncate variant="large">
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
}

const LoadComponentConfig: FC<{
  componentId: string
  componentConfigId: string
  orgId: string
}> = async ({ componentId, componentConfigId, orgId }) => {
  const { data: componentConfig, error } = await nueQueryData<TComponentConfig>(
    {
      orgId,
      path: `components/${componentId}/configs/${componentConfigId}`,
    }
  )
  return error ? (
    <Text>{error?.error}</Text>
  ) : (
    <ComponentConfiguration config={componentConfig} />
  )
}
