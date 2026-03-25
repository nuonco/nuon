'use client'

import { Plan } from '@/components/approvals/Plan'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { SSELogs } from '@/components/log-stream/SSELogs'
import { useOrg } from '@/hooks/use-org'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider'
import type { TWorkflowStep, TSandboxRun } from '@/types'
import { useQuery } from '@tanstack/react-query'
import { getInstallSandboxRun } from '@/lib'
import {
  SandboxRunApply,
  SandboxRunApplySkeleton,
  SandboxRunLogsSkeleton,
} from './SandboxRunApply'

interface ISandboxRunStepDetails {
  step?: TWorkflowStep
}

export const SandboxRunStepDetails = ({ step }: ISandboxRunStepDetails) => {
  const { org } = useOrg()

  const { data: sandboxRun, isLoading } = useQuery<TSandboxRun>({
    queryKey: ['sandbox-run', org?.id, step?.step_target_id],
    queryFn: () =>
      getInstallSandboxRun({ orgId: org.id, runId: step!.step_target_id }),
    enabled: !!org?.id && !!step?.step_target_id,
  })

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-4">
        <Text variant="base" weight="strong">
          Sandox run
        </Text>

        <Text variant="subtext">
          <Link href={`/${org.id}/installs/${step.owner_id}/sandbox`}>
            View sandbox <Icon variant="CaretRight" />
          </Link>
        </Text>

        <Text variant="subtext">
          <Link
            href={`/${org.id}/installs/${step.owner_id}/sandbox/runs/${step.step_target_id}`}
          >
            View run logs <Icon variant="CaretRight" />
          </Link>
        </Text>

      </div>
      <div className="flex flex-wrap gap-x-4 gap-y-1 items-center">
        {sandboxRun?.created_at ? (
          <Text theme="info" className="!flex items-center gap-1">
            <Icon variant="CalendarBlankIcon" />
            <Time variant="subtext" time={sandboxRun.created_at} />
          </Text>
        ) : null}
        {sandboxRun?.created_at ? (
          <Text theme="info" className="!flex items-center gap-1">
            <Icon variant="TimerIcon" />
            <Duration variant="subtext" beginTime={sandboxRun.created_at} endTime={sandboxRun.updated_at} />
          </Text>
        ) : null}
        {sandboxRun?.role ? (
          <Text theme="info" className="!flex items-center gap-1">
            <Icon variant="FileLockIcon" />
            <Text variant="subtext">{sandboxRun.role}</Text>
          </Text>
        ) : null}
      </div>

      {step?.execution_type === 'approval' ? (
        <Tabs
          tabs={{
            plan: (
              <div className="mt-4">
                <Plan step={step} />
              </div>
            ),
            logs: sandboxRun?.log_stream ? (
              <LogStreamProvider
                shouldPoll
                logStreamId={sandboxRun.log_stream.id}
              >
                <UnifiedLogsProvider>
                  <LogViewerProvider>
                    <SSELogs />
                  </LogViewerProvider>
                </UnifiedLogsProvider>
              </LogStreamProvider>
            ) : (
              <SandboxRunLogsSkeleton />
            ),
          }}
        />
      ) : isLoading && !sandboxRun ? (
        <div className="flex flex-col gap-4">
          <SandboxRunApplySkeleton />
          <SandboxRunLogsSkeleton />
        </div>
      ) : (
        <SandboxRunApply step={step} sandboxRun={sandboxRun} />
      )}
    </div>
  )
}
