import type { Metadata } from 'next'
import { BuildHeader } from '@/components/builds/BuildHeader'
import { AsyncBoundary } from '@/components/common/AsyncBoundary'
import { BackToTop } from '@/components/common/BackToTop'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { BuildProvider } from '@/providers/build-provider'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { getApp, getComponentBuild, getComponent, getOrg } from '@/lib'

import { Logs, LogsError, LogsSkeleton } from './logs'
import { RefreshLogStream } from './refresh-log-stream'

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
          <div>
            {build?.log_stream ? (
              <LogStreamProvider
                initLogStream={build?.log_stream}
                shouldPoll={build?.log_stream?.open}
              >
                <AsyncBoundary
                  errorFallback={<LogsError />}
                  loadingFallback={<LogsSkeleton />}
                >
                  <Logs
                    logStreamId={build?.log_stream?.id}
                    logStreamOpen={build?.log_stream?.open}
                    orgId={orgId}
                  />
                </AsyncBoundary>
              </LogStreamProvider>
            ) : (
              <RefreshLogStream />
            )}
          </div>

          <BackToTop containerId={containerId} />
        </PageSection>
      </BuildProvider>
    </>
  )
}
