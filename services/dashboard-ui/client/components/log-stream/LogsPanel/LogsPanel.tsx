import type { ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { SSELogs } from '@/components/log-stream/SSELogs'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import type { TLogStream } from '@/types'

export interface ILogsPanel {
  children?: ReactNode
  filterClassName?: string
  logStream?: TLogStream
}

export const LogsPanel = ({
  children,
  filterClassName = 'top-0',
  logStream,
}: ILogsPanel) => {
  if (!logStream?.id) {
    return (
      <div className="flex flex-col items-center gap-4 p-12">
        <Text variant="base" weight="strong">
          Waiting on log stream
        </Text>
        <Text variant="body" theme="neutral">
          Logs will appear here once the runner starts.
        </Text>
        <Button variant="ghost" onClick={() => window.location.reload()}>
          <Icon variant="ArrowClockwiseIcon" />
          Refresh page
        </Button>
      </div>
    )
  }

  return (
    <LogStreamProvider logStreamId={logStream.id}>
      <LogViewerProvider>
        {children ?? <SSELogs filterClassName={filterClassName} />}
      </LogViewerProvider>
    </LogStreamProvider>
  )
}
