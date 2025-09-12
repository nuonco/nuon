'use client'

import { Empty } from '@/components/Empty'
import { Timeline } from '@/components/Timeline'
import { ToolTip } from '@/components/ToolTip'
import { Text, Truncate } from '@/components/Typography'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useQueryParams } from '@/hooks/use-query-params'
import type { TSandboxRun, TPaginationParams } from '@/types'

export interface ISandboxHistory extends IPollingProps, TPaginationParams {
  installId: string
  initSandboxRuns: Array<TSandboxRun>
  orgId: string
}

export const SandboxHistory = ({
  installId,
  initSandboxRuns,
  pollInterval = 5000,
  shouldPoll = false,
  orgId,
  offset,
  limit,
}: ISandboxHistory) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const params = useQueryParams({ offset, limit })
  const { data: sandboxRuns } = usePolling<TSandboxRun[]>({
    dependencies: [params],
    initData: initSandboxRuns,
    path: `/api/orgs/${org.id}/installs/${install.id}/sandbox/runs${params}`,
    pollInterval,
    shouldPoll,
  })

  return (
    <Timeline
      emptyContent={
        <Empty
          emptyTitle="No runs yet"
          emptyMessage="Waiting on sandbox runs."
          variant="history"
          isSmall
        />
      }
      events={sandboxRuns.map((run, i) => ({
        id: run.id,
        status: run?.status_v2?.status || run?.status,
        underline: (
          <div>
            <Text>
              <Text variant="mono-12">
                <ToolTip tipContent={run.id}>
                  <Truncate variant="small">{run.id}</Truncate>
                </ToolTip>
              </Text>
              <>
                /{' '}
                {run?.run_type.length >= 12 ? (
                  <ToolTip tipContent={run?.run_type} alignment="right">
                    <Truncate variant="small">{run?.run_type}</Truncate>
                  </ToolTip>
                ) : (
                  run.run_type
                )}
              </>
            </Text>
            {run?.created_by ? (
              <Text className="text-cool-grey-600 dark:text-white/70 !text-[10px]">
                Run by: {run?.created_by?.email}
              </Text>
            ) : null}
          </div>
        ),
        time: run.updated_at,
        href:
          (run?.status_v2?.status &&
            (run?.status_v2?.status as string) !== 'queued') ||
          (run?.status && run?.status !== 'queued')
            ? `/${orgId}/installs/${installId}/sandbox/${run.id}`
            : null,
        isMostRecent: i === 0,
      }))}
    />
  )
}
