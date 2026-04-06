import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { SSELogs, LogsSkeleton } from '@/components/log-stream/SSELogs'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageSection } from '@/components/layout/PageSection'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import { RunnerProvider } from '@/providers/runner-provider'
import { useOrg } from '@/hooks/use-org'
import { getRunnerProcess } from '@/lib'

export const ProcessSystemLogs = () => {
  const { org } = useOrg()
  const { processId } = useParams<{ processId: string }>()
  const runnerId = org?.runner_group?.runners?.[0]?.id

  const { data: process, isLoading } = useQuery({
    queryKey: ['runner-process', org?.id, runnerId, processId],
    queryFn: () =>
      getRunnerProcess({
        orgId: org.id,
        runnerId: runnerId!,
        processId: processId!,
      }),
    enabled: !!org?.id && !!runnerId && !!processId,
  })

  const breadcrumbs = (
    <>
      <PageTitle title={`System logs | ${org?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org?.name },
          { path: `/${org.id}/runner`, text: 'Build runner' },
          {
            path: `/${org.id}/runner/processes/${processId}/logs`,
            text: 'System logs',
          },
        ]}
      />
    </>
  )

  if (isLoading || !process) {
    return (
      <PageLayout>
        {breadcrumbs}
        <PageContent>
          <LogsSkeleton />
        </PageContent>
      </PageLayout>
    )
  }

  if (!process.log_stream_id) {
    return (
      <PageLayout>
        {breadcrumbs}
        <PageContent>
          <div className="flex flex-col items-center gap-4 p-12">
            <Text variant="base" weight="strong">
              No log stream available
            </Text>
            <Text variant="body" theme="neutral">
              This process does not have a log stream configured.
            </Text>
            <Button variant="ghost" onClick={() => window.history.back()}>
              <Icon variant="ArrowLeftIcon" />
              Back to runner
            </Button>
          </div>
        </PageContent>
      </PageLayout>
    )
  }

  return (
    <RunnerProvider runnerId={runnerId!}>
      <PageLayout>
        {breadcrumbs}
        <PageHeader>
          <Text variant="h3" weight="strong" level={1}>
            System logs — {process.type} process
          </Text>
        </PageHeader>
        <PageSection>
          <LogStreamProvider logStreamId={process.log_stream_id} shouldPoll>
            <UnifiedLogsProvider>
              <LogViewerProvider>
                <SSELogs />
              </LogViewerProvider>
            </UnifiedLogsProvider>
          </LogStreamProvider>
        </PageSection>
      </PageLayout>
    </RunnerProvider>
  )
}
