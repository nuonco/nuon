import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { BuildHeader } from '@/components/builds/BuildHeader'
import { SSELogs } from '@/components/log-stream/SSELogs'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useBuild } from '@/hooks/use-build'
import { useOrg } from '@/hooks/use-org'
import { getComponent } from '@/lib'
import { BuildProvider } from '@/providers/build-provider'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider'
import type { TComponent } from '@/types'

const BuildDetailInner = ({ component }: { component: TComponent | undefined }) => {
  const { build } = useBuild()

  if (!build) return null

  return (
    <>
      <BuildHeader component={component as TComponent} />
      {build?.status_v2?.metadata?.duplicate_build ? (
        <Banner theme="warn" className="mx-6 mt-4">
          <div className="flex flex-col">
            <Text weight="strong" variant="base">
              Duplicate build
            </Text>
            <Text theme="neutral">
              This build was triggered against the same commit and config as a
              previous build. Push new changes to your branch to create a unique
              build.
            </Text>
          </div>
        </Banner>
      ) : null}
      <PageSection>
        {build?.log_stream ? (
          <LogStreamProvider
            logStreamId={build.log_stream.id}
            shouldPoll={build.log_stream.open}
          >
            <UnifiedLogsProvider>
              <LogViewerProvider>
                <SSELogs />
              </LogViewerProvider>
            </UnifiedLogsProvider>
          </LogStreamProvider>
        ) : (
          <div className="flex flex-col items-center gap-4 p-12">
            <Text variant="base" weight="strong">
              Waiting on log stream
            </Text>
            <Text variant="body" theme="neutral">
              Logs will appear here once the build runner starts.
            </Text>
            <Button
              variant="ghost"
              onClick={() => window.location.reload()}
            >
              <Icon variant="ArrowClockwiseIcon" />
              Refresh Page
            </Button>
          </div>
        )}
      </PageSection>
    </>
  )
}

export const BuildDetail = () => {
  const { componentId, buildId } = useParams()
  const { org } = useOrg()
  const { app } = useApp()

  const { data: component } = useQuery({
    queryKey: ['component', org?.id, app?.id, componentId],
    queryFn: () => getComponent({ orgId: org.id, componentId: componentId! }),
    enabled: !!org?.id && !!app?.id && !!componentId,
  })

  return (
    <div className="flex flex-col flex-1">
      <PageTitle title={`Build | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/components`, text: 'Components' },
          { path: `/${org?.id}/apps/${app?.id}/components/${componentId}`, text: component?.name },
          { path: `/${org?.id}/apps/${app?.id}/components/${componentId}/builds/${buildId}`, text: 'Build' },
        ]}
      />
      <BuildProvider buildId={buildId!} componentId={componentId!} shouldPoll>
        <BuildDetailInner component={component} />
      </BuildProvider>
    </div>
  )
}
