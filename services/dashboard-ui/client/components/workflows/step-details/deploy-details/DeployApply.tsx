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
import type { TInstallDeploy } from '@/types'

export const DeployApply = ({
  initDeploy: deploy,
}: {
  initDeploy?: TInstallDeploy
}) => {
  return (
    <>
      {!deploy ? (
        <div className="flex flex-col gap-4">
          <div className="flex items-start gap-6">
            <LabeledStatus label="Status" loading />
            <LabeledValue label="Deploy ID" loading />
          </div>
          <LogsSkeleton />
        </div>
      ) : (
        <div className="flex flex-col gap-4">
          <div className="flex items-start gap-6">
            <LabeledStatus
              label="Status"
              statusProps={{
                status: deploy?.status_v2?.status,
              }}
              tooltipProps={{
                position: 'right',
                tipContent: (
                  <Text nowrap variant="subtext">
                    {deploy?.status_v2?.status_human_description}
                  </Text>
                ),
              }}
            />
            <LabeledValue label="Deploy ID">
              <ID>{deploy?.id}</ID>
            </LabeledValue>
          </div>

          {deploy?.log_stream ? (
            <LogStreamProvider logStreamId={deploy?.log_stream?.id}>
              <LogViewerProvider>
                <Tabs
                  tabs={{
                    logs: <SSELogs />,
                    trace: (
                      <TraceView
                        logStreamId={deploy.log_stream.id}
                        shouldPoll={deploy.log_stream.open}
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
