'use client'

import { ID } from '@/components/common/ID'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Skeleton } from '@/components/common/Skeleton'
import { Logs, LogsSkeleton } from '@/components/log-stream/Logs'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogsProvider } from '@/providers/logs-provider'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import type { TInstallDeploy, TOTELLog, TWorkflowStep } from '@/types'

export const DeployApply = ({
  initDeploy: deploy,
}: {
  initDeploy: TInstallDeploy
}) => {
  const { org } = useOrg()

  const { data: logs, isLoading: isLoadingLogs } = useQuery<TOTELLog[]>({
    initData: [],
    path: `/api/orgs/${org.id}/log-streams/${deploy?.log_stream?.id}/logs`,
  })

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
                tipContent: deploy?.status_v2?.status_human_description,
              }}
            />
            <LabeledValue label="Deploy ID">
              <ID>{deploy?.id}</ID>
            </LabeledValue>
          </div>

          {deploy?.log_stream ? (
            <LogStreamProvider shouldPoll initLogStream={deploy?.log_stream}>
              {isLoadingLogs ? (
                <DeployLogsSkeleton />
              ) : (
                <LogsProvider initLogs={logs}>
                  <Logs />
                </LogsProvider>
              )}
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

const DeployLogsSkeleton = () => {
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
