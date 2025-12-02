import type { Metadata } from 'next'
import { Suspense } from 'react'
import { BackLink } from '@/components/common/BackLink'
import { BackToTop } from '@/components/common/BackToTop'
import { ErrorBoundary } from "@/components/common/ErrorBoundary"
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import {
  getAppById,
  getComponentBuildById,
  getComponentById,
  getOrgById,
} from '@/lib'
import { Logs, LogsError, LogsSkeleton } from './logs'

// NOTE: old layout stuff
import { CalendarBlankIcon, TimerIcon } from '@phosphor-icons/react/dist/ssr'
import {
  ComponentConfigType,
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
import { BuildDetails } from '@/components/old/Components/BuildDetails'

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
    title: `SSE build logs | ${build?.component_name} | Nuon`,
  }
}

export default async function AppComponentBuildPage({ params }) {
  const {
    ['org-id']: orgId,
    ['app-id']: appId,
    ['component-id']: componentId,
    ['build-id']: buildId,
  } = await params

  const [{ data: app }, { data: build }, { data: component }, { data: org }] =
    await Promise.all([
      getAppById({ appId, orgId }),
      getComponentBuildById({ componentId, buildId, orgId }),
      getComponentById({ componentId, orgId }),
      getOrgById({ orgId }),
    ])

  const containerId = 'component-build-page'
  return (
    <PageSection className="!p-0 !gap-0" id={containerId} isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/apps`,
            text: 'Apps',
          },
          {
            path: `/${orgId}/apps/${appId}`,
            text: app?.name,
          },
          {
            path: `/${orgId}/apps/${appId}/components`,
            text: 'Components',
          },
          {
            path: `/${orgId}/apps/${appId}/components/${componentId}`,
            text: component?.name,
          },
          {
            path: `/${orgId}/apps/${appId}/components/${componentId}/builds/${buildId}`,
            text: 'Build',
          },
        ]}
      />
      {/* old page layout */}
      <div className="p-6 border-b flex justify-between">
        <HeadingGroup>
          <BackLink className="mb-6" />
          <span className="flex items-center gap-2">
            <ComponentConfigType configType={component?.type} isIconOnly />
            <Text variant="base" weight="strong">
              {build?.component_name}
            </Text>
          </span>
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



          {build?.log_stream ? (
            <Section heading="Build logs">
              <LogStreamProvider
                initLogStream={build?.log_stream}
                shouldPoll={build?.log_stream?.open}
              >
                <ErrorBoundary fallback={<LogsError />}>
                  <Suspense fallback={<LogsSkeleton />}>
                    <Logs
                      logStreamId={build?.log_stream?.id}
                      logStreamOpen={build?.log_stream?.open}
                      orgId={orgId}
                    />
                  </Suspense>
                </ErrorBoundary>
              </LogStreamProvider>
            </Section>
          ) : (
            <Section heading="Build logs">
              <OldText>Waiting on log stream</OldText>
            </Section>
          )}


      {/* old page layout */}
      <BackToTop containerId={containerId} />
    </PageSection>
  )
}
