import type { Metadata } from 'next'
import { Suspense } from 'react'
import { BuildHeader } from '@/components/builds/BuildHeader'
import { BackLink } from '@/components/common/BackLink'
import { BackToTop } from '@/components/common/BackToTop'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { BuildProvider } from '@/providers/build-provider'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import {
  getApp,
  getComponentBuild,
  getComponent,
  getOrg,
} from '@/lib'
import { ComponentConfig } from './config'
import { Logs, LogsError, LogsSkeleton } from './logs'
import { RefreshLogStream } from './refresh-log-stream'

// NOTE: old layout stuff
import { ErrorBoundary } from 'react-error-boundary'
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
  const { data: build } = await getComponentBuild({
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

  const [{ data: app }, { data: build }, { data: component }, { data: org }] =
    await Promise.all([
      getApp({ appId, orgId }),
      getComponentBuild({ componentId, buildId, orgId }),
      getComponent({ componentId, orgId }),
      getOrg({ orgId }),
    ])

  const containerId = 'component-build-page'
  return (
    <>
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
      <BuildProvider initBuild={build}>
        <BuildHeader component={component} />
        <PageSection id={containerId} isScrollable>
          {/* old page layout */}
          <div>
            {build?.log_stream ? (
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
            ) : (
              <RefreshLogStream />
            )}
          </div>
          {/* old page layout */}

          <BackToTop containerId={containerId} />
        </PageSection>
      </BuildProvider>
    </>
  )
}
