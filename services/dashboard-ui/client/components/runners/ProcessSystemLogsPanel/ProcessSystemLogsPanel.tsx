import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { SSELogs, LogsSkeleton } from '@/components/log-stream/SSELogs'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import { RunnerProvider } from '@/providers/runner-provider'
import type { TRunnerProcess } from '@/types'

interface IProcessSystemLogsPanel {
  process?: TRunnerProcess
  isLoading?: boolean
  runnerId?: string
}

export const ProcessSystemLogsPanel = ({
  process,
  isLoading,
  runnerId,
}: IProcessSystemLogsPanel) => {
  if (isLoading || !process) {
    return (
      <PageSection>
        <LogsSkeleton />
      </PageSection>
    )
  }

  if (!process.log_stream_id) {
    return (
      <PageSection>
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
      </PageSection>
    )
  }

  return (
    <RunnerProvider runnerId={runnerId!}>
      <PageHeader>
        <Text variant="h3" weight="strong" level={1}>
          System logs — {process.type} process
        </Text>
      </PageHeader>
      <PageSection>
        <LogStreamProvider logStreamId={process.log_stream_id}>
          <LogViewerProvider>
            <SSELogs />
          </LogViewerProvider>
        </LogStreamProvider>
      </PageSection>
    </RunnerProvider>
  )
}
