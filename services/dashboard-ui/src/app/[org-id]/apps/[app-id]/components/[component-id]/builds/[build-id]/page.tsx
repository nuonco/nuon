import type { Metadata } from 'next'
import { Suspense } from 'react'
import { BackLink } from '@/components/common/BackLink'
import { BackToTop } from '@/components/common/BackToTop'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { PageSection } from '@/components/layout/PageSection'
import { Text } from '@/components/common/Text'
import { BuildDetails } from '@/components/Components/BuildDetails'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { getAppById, getComponentBuildById, getOrgById } from '@/lib'
import { ComponentConfig } from './config'
import { Logs, LogsError, LogsSkeleton } from './logs'

// NOTE: old layout stuff
import { ErrorBoundary } from 'react-error-boundary'
import { CalendarBlankIcon, TimerIcon } from '@phosphor-icons/react/dist/ssr'
import {
  DashboardContent,
  Duration,
  ErrorFallback,
  Loading,
  LogStreamProvider as OldLogStreamProvider,
  OperationLogsSection,
  Section,
  Time,
  Text as OldText,
  ToolTip,
  Truncate,
} from '@/components'

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

export default async function AppComponentBuildPage({ params }) {
  const {
    ['org-id']: orgId,
    ['app-id']: appId,
    ['component-id']: componentId,
    ['build-id']: buildId,
  } = await params

  const [{ data: app }, { data: build }, { data: org }] = await Promise.all([
    getAppById({ appId, orgId }),
    getComponentBuildById({ componentId, buildId, orgId }),
    getOrgById({ orgId }),
  ])

  const containerId = 'component-build-page'
  return org?.features?.['stratus-layout'] ? (
    <PageSection className="!p-0 !gap-0" id={containerId} isScrollable>
      {/* old page layout */}
      <div className="p-6 border-b flex justify-between">
        <HeadingGroup>
          <BackLink className="mb-6" />
          <Text variant="base" weight="strong">
            {build?.component_name}
          </Text>
          <ID>{buildId}</ID>
          <div className="flex gap-8 items-center justify-start mt-2">
            <Text className="!flex items-center gap-1">
              <CalendarBlankIcon />
              <Time time={build.created_at} />
            </Text>
            <Text className="!flex items-center gap-1">
              <TimerIcon />
              <Duration
                beginTime={build.created_at}
                endTime={build.updated_at}
              />
            </Text>
          </div>
        </HeadingGroup>

        <BuildDetails initBuild={build} shouldPoll />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="md:col-span-8">
          {build?.log_stream ? (
            <Section heading="Build logs">
              <LogStreamProvider
                initLogStream={build?.log_stream}
                shouldPoll={build?.log_stream?.open}
              >
                <ErrorBoundary fallback={<LogsError />}>
                  <Suspense fallback={<LogsSkeleton />}>
                    <Logs logStreamId={build?.log_stream?.id} orgId={orgId} />
                  </Suspense>
                </ErrorBoundary>
              </LogStreamProvider>
            </Section>
          ) : (
            <Section heading="Build logs">
              <OldText>Waiting on log stream</OldText>
            </Section>
          )}
        </div>
        <div className="divide-y flex flex-col md:col-span-4">
          {build.vcs_connection_commit && (
            <Section className="flex-initial" heading="Commit details">
              <div className="flex gap-6 items-start justify-start">
                <span className="flex flex-col gap-2">
                  <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                    SHA
                  </OldText>
                  <ToolTip tipContent={build.vcs_connection_commit?.sha}>
                    <OldText
                      className="truncate text-ellipsis w-16"
                      variant="mono-12"
                    >
                      {build.vcs_connection_commit?.sha}
                    </OldText>
                  </ToolTip>
                </span>

                {build.vcs_connection_commit?.author_name !== '' && (
                  <span className="flex flex-col gap-2">
                    <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                      Author
                    </OldText>
                    <OldText>
                      {build.vcs_connection_commit?.author_name}
                    </OldText>
                  </span>
                )}

                <span className="flex flex-col gap-2">
                  <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                    Message
                  </OldText>
                  <OldText>
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
                  </OldText>
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
      {/* old page layout */}
      <BackToTop containerId={containerId} />
    </PageSection>
  ) : (
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
          <OldText>
            <CalendarBlankIcon />
            <Time time={build.created_at} />
          </OldText>
          <OldText>
            <TimerIcon />
            <Duration beginTime={build.created_at} endTime={build.updated_at} />
          </OldText>
        </div>
      }
      statues={<BuildDetails initBuild={build} shouldPoll />}
    >
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="md:col-span-8">
          {build?.log_stream ? (
            <OldLogStreamProvider initLogStream={build?.log_stream}>
              <OperationLogsSection heading="Build logs" />
            </OldLogStreamProvider>
          ) : (
            <Section heading="Build logs">
              <OldText>Waiting on log stream</OldText>
            </Section>
          )}
        </div>
        <div className="divide-y flex flex-col md:col-span-4">
          {build.vcs_connection_commit && (
            <Section className="flex-initial" heading="Commit details">
              <div className="flex gap-6 items-start justify-start">
                <span className="flex flex-col gap-2">
                  <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                    SHA
                  </OldText>
                  <ToolTip tipContent={build.vcs_connection_commit?.sha}>
                    <OldText
                      className="truncate text-ellipsis w-16"
                      variant="mono-12"
                    >
                      {build.vcs_connection_commit?.sha}
                    </OldText>
                  </ToolTip>
                </span>

                {build.vcs_connection_commit?.author_name !== '' && (
                  <span className="flex flex-col gap-2">
                    <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                      Author
                    </OldText>
                    <OldText>
                      {build.vcs_connection_commit?.author_name}
                    </OldText>
                  </span>
                )}

                <span className="flex flex-col gap-2">
                  <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                    Message
                  </OldText>
                  <OldText>
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
                  </OldText>
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
