import { Text } from '@/components/common/Text'
import { SSELogs, LogsSkeleton } from '@/components/log-stream/SSELogs'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider'
import type { TWorkflowStep } from '@/types'

export interface ISyncSecretsStepDetails {
  step?: TWorkflowStep
}

export const SyncSecretsStepDetails = ({ step }: ISyncSecretsStepDetails) => {
  const logStreamId = step?.log_stream?.id

  return (
    <div className="flex flex-col gap-4">
      <Text variant="base" weight="strong">
        Sync secrets
      </Text>

      {logStreamId ? (
        <LogStreamProvider shouldPoll logStreamId={logStreamId}>
          <UnifiedLogsProvider>
            <LogViewerProvider>
              <SSELogs />
            </LogViewerProvider>
          </UnifiedLogsProvider>
        </LogStreamProvider>
      ) : (
        <div className="flex flex-col divide-y">
          <LogsSkeleton />
        </div>
      )}
    </div>
  )
}
