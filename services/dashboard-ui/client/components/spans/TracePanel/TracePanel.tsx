import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { TraceView } from '@/components/spans/TraceView'
import { useOrg } from '@/hooks/use-org'
import { getLogStreamSpans } from '@/lib'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider'
import type { TLogStream } from '@/types'

export interface ITracePanel {
  logStream?: TLogStream
}

// Wrapper around TraceView that handles the missing-log-stream empty state and
// the standard provider stack. Used by every Trace tab / Trace panel so they
// stay in sync.
//
// Fetches spans HERE (rather than inside TraceView) so they can be threaded
// down into LogViewerProvider for parent-span aggregation in the cross-link.
// Without that, clicking a parent step span (e.g. step.execute) would only
// match logs whose span_id is literally step.execute — but most logs are
// emitted from inside child op spans (terraform.plan, helm.upgrade, …) so
// the user would see "0 logs" until they clicked a leaf.
export const TracePanel = ({ logStream }: ITracePanel) => {
  if (!logStream?.id) {
    return (
      <div className="flex flex-col items-center gap-4 p-12">
        <Text variant="base" weight="strong">
          Waiting on log stream
        </Text>
        <Text variant="body" theme="neutral">
          The trace will appear here once the runner starts emitting spans.
        </Text>
        <Button variant="ghost" onClick={() => window.location.reload()}>
          <Icon variant="ArrowClockwiseIcon" />
          Refresh page
        </Button>
      </div>
    )
  }

  return <TracePanelContent logStream={logStream} />
}

const TracePanelContent = ({ logStream }: { logStream: TLogStream }) => {
  const { org } = useOrg()
  const shouldPoll = logStream.open
  const { data: spans, isLoading } = useQuery({
    queryKey: ['log-stream-spans', org?.id, logStream.id],
    queryFn: () => getLogStreamSpans({ orgId: org.id, logStreamId: logStream.id }),
    enabled: !!org?.id && !!logStream.id,
    refetchInterval: shouldPoll ? 5000 : false,
  })

  return (
    <div className="flex flex-col gap-6 flex-auto min-h-0">
      <LogStreamProvider logStreamId={logStream.id} shouldPoll={shouldPoll}>
        <UnifiedLogsProvider>
          <LogViewerProvider spans={spans}>
            <TraceView
              logStreamId={logStream.id}
              shouldPoll={shouldPoll}
              spans={spans}
              spansLoading={isLoading}
            />
          </LogViewerProvider>
        </UnifiedLogsProvider>
      </LogStreamProvider>
    </div>
  )
}
