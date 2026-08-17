import { ID } from '@/components/common/ID'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { LogsSkeleton } from '@/components/log-stream/SSELogs'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { SSELogs } from '@/components/log-stream/SSELogs'
import { TraceView } from '@/components/spans/TraceView'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import type { TWorkflowStep, TSandboxRun } from '@/types'

export const SandboxRunApply = ({
  sandboxRun,
}: {
  step?: TWorkflowStep
  sandboxRun?: TSandboxRun
}) => {
  return (
    <>
      {!sandboxRun ? (
        <div className="flex flex-col gap-4">
          <div className="flex items-start gap-6">
            <LabeledStatus label="Status" loading />
            <LabeledValue label="Sandbox Run ID" loading />
          </div>
          <LogsSkeleton />
        </div>
      ) : (
        <div className="flex flex-col gap-4">
          <div className="flex items-start gap-6">
            <LabeledStatus
              label="Status"
              statusProps={{
                status: sandboxRun?.status_v2?.status,
              }}
              tooltipProps={{
                position: 'right',
                tipContent: (
                  <Text nowrap variant="subtext">
                    {sandboxRun?.status_v2?.status_human_description}
                  </Text>
                ),
              }}
            />
            <LabeledValue label="Sandbox Run ID">
              <ID>{sandboxRun?.id}</ID>
            </LabeledValue>
          </div>

          {sandboxRun?.log_stream ? (
            <LogStreamProvider logStreamId={sandboxRun?.log_stream?.id}>
              <LogViewerProvider>
                <Tabs
                  tabs={{
                    logs: <SSELogs />,
                    trace: (
                      <TraceView
                        logStreamId={sandboxRun.log_stream.id}
                        shouldPoll={sandboxRun.log_stream.open}
                      />
                    ),
                  }}
                />
              </LogViewerProvider>
            </LogStreamProvider>
          ) : null}
        </div>
      )}
    </>
  )
}
