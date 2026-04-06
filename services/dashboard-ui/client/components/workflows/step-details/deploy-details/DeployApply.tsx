import { ID } from '@/components/common/ID'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { LogsSkeleton } from '@/components/log-stream/SSELogs'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { SSELogs } from '@/components/log-stream/SSELogs'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import type { TInstallDeploy } from '@/types'

export const DeployApply = ({
  initDeploy: deploy,
}: {
  initDeploy: TInstallDeploy
}) => {
  return (
    <>
      {!deploy ? (
        <div className="flex flex-col gap-4">
          <DeployApplySkeleton />
          <DeployLogsSkeleton />
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
            <LogStreamProvider shouldPoll logStreamId={deploy?.log_stream?.id}>
              <UnifiedLogsProvider>
                <LogViewerProvider>
                  <SSELogs />
                </LogViewerProvider>
              </UnifiedLogsProvider>
            </LogStreamProvider>
          ) : null}
        </div>
      )}
    </>
  )
}

export const DeployApplySkeleton = () => {
  return (
    <div className="flex items-start gap-6">
      <LabeledValue label={<Skeleton height="17px" width="34px" />}>
        <Skeleton height="23px" width="75px" />
      </LabeledValue>
    </div>
  )
}

export const DeployLogsSkeleton = () => {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Skeleton height="36px" width="320px" />
          <Skeleton height="17px" width="85px" />
        </div>

        <div className="flex items-center gap-4">
          <Skeleton height="32px" width="86px" />
          <Skeleton height="32px" width="135px" />
          <Skeleton height="32px" width="140px" />
        </div>
      </div>
      <div>
        <LogsSkeleton />
      </div>
    </div>
  )
}
