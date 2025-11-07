'use client'

import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Skeleton } from '@/components/common/Skeleton'
import { Logs, LogsSkeleton } from '@/components/log-stream/Logs'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogsProvider } from '@/providers/logs-provider'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import type { TOTELLog, TWorkflowStep, TSandboxRun } from '@/types'

export const SandboxRunApply = ({
  step,
  sandboxRun,
}: {
  step: TWorkflowStep
  sandboxRun: TSandboxRun
}) => {
  const { org } = useOrg()

  const { data: logs, isLoading: isLoadingLogs } = useQuery<TOTELLog[]>({
    initData: [],
    path: `/api/orgs/${org.id}/log-streams/${sandboxRun?.log_stream?.id}/logs`,
  })

  return (
    <>
      {!sandboxRun ? (
        <div className="flex flex-col gap-4">
          <SandboxRunApplySkeleton />
          <SandboxRunLogsSkeleton />
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
                position: 'top',
                tipContent: sandboxRun?.status_v2?.status_human_description,
              }}
            />
          </div>

          {sandboxRun?.log_stream ? (
            <LogStreamProvider
              shouldPoll
              initLogStream={sandboxRun?.log_stream}
            >
              {isLoadingLogs ? (
                <SandboxRunLogsSkeleton />
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

export const SandboxRunApplySkeleton = () => {
  return (
    <div className="flex items-start gap-6">
      <LabeledValue label={<Skeleton height="17px" width="34px" />}>
        <Skeleton height="23px" width="75px" />
      </LabeledValue>
    </div>
  )
}

export const SandboxRunLogsSkeleton = () => {
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
