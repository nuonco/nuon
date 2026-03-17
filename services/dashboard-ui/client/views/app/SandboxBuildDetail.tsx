import { useParams } from 'react-router'
import { BackLink } from '@/components/common/BackLink'
import { Button } from '@/components/common/Button'
import { Duration } from '@/components/common/Duration'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { BackToTop } from '@/components/common/BackToTop'
import { SSELogs } from '@/components/log-stream/SSELogs'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSandboxBuild } from '@/hooks/use-sandbox-build'
import { SandboxBuildProvider } from '@/providers/sandbox-build-provider'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider'

const CONTAINER_ID = 'sandbox-build-detail-page'

const SandboxBuildDetailInner = () => {
  const { build } = useSandboxBuild()

  if (!build) return null

  return (
    <>
      <header className="p-6 border-b flex justify-between">
        <HeadingGroup>
          <BackLink className="mb-6" />
          <div className="flex flex-col gap-1">
            <Text variant="base" weight="strong">
              Sandbox build
            </Text>
            <ID>{build?.id}</ID>
          </div>
          <div className="flex gap-8 items-center justify-start mt-2">
            <Text theme="info" className="!flex items-center gap-1">
              <Icon variant="CalendarBlankIcon" />
              <Time variant="subtext" time={build.created_at} />
            </Text>
            <Text theme="info" className="!flex items-center gap-1">
              <Icon variant="TimerIcon" />
              <Duration
                variant="subtext"
                beginTime={build.created_at}
                endTime={build.updated_at}
              />
            </Text>
          </div>
        </HeadingGroup>

        <div className="flex flex-col gap-6">
          <div className="flex gap-6 items-start justify-start">
            <LabeledStatus
              label="Status"
              statusProps={{
                status: build?.status_v2?.status ?? build?.status,
              }}
              tooltipProps={{
                tipContentClassName: 'w-fit',
                tipContent: (
                  <Text className="!text-nowrap" variant="subtext">
                    {build?.status_v2?.status_human_description ?? build?.status_description}
                  </Text>
                ),
              }}
            />
            <LabeledValue label="Built by">
              <Text variant="body">{build?.created_by?.email ?? '—'}</Text>
            </LabeledValue>
          </div>
        </div>
      </header>

      <PageSection id={CONTAINER_ID} isScrollable>
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
        <BackToTop containerId={CONTAINER_ID} />
      </PageSection>
    </>
  )
}

export const SandboxBuildDetail = () => {
  const { buildId } = useParams()
  const { org } = useOrg()
  const { app } = useApp()

  return (
    <div className="flex flex-col flex-1 overflow-hidden">
      <PageTitle title={`Sandbox Build | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/sandbox`, text: 'Sandbox' },
          {
            path: `/${org?.id}/apps/${app?.id}/sandbox/builds/${buildId}`,
            text: 'Build',
          },
        ]}
      />
      <SandboxBuildProvider buildId={buildId!} shouldPoll>
        <SandboxBuildDetailInner />
      </SandboxBuildProvider>
    </div>
  )
}