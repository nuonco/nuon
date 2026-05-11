import { Plan } from '@/components/approvals/Plan'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { SSELogs } from '@/components/log-stream/SSELogs'
import { TraceView } from '@/components/spans/TraceView'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider'
import type { TWorkflowStep, TSandboxRun } from '@/types'
import {
  SandboxRunApply,
  SandboxRunApplySkeleton,
  SandboxRunLogsSkeleton,
} from '../SandboxRunApply'

export interface ISandboxRunStepDetails {
  step?: TWorkflowStep
  orgId: string
  sandboxRun?: TSandboxRun
  isLoading: boolean
}

export const SandboxRunStepDetails = ({
  step,
  orgId,
  sandboxRun,
  isLoading,
}: ISandboxRunStepDetails) => {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-4">
        <Text variant="base" weight="strong">
          Sandox run
        </Text>

        <Text variant="subtext">
          <Link href={`/${orgId}/installs/${step?.owner_id}/sandbox`}>
            View sandbox <Icon variant="CaretRight" />
          </Link>
        </Text>

        <Text variant="subtext">
          <Link
            href={`/${orgId}/installs/${step?.owner_id}/sandbox/runs/${step?.step_target_id}`}
          >
            View run logs <Icon variant="CaretRight" />
          </Link>
        </Text>

      </div>
      <div className="flex flex-wrap gap-x-4 gap-y-1 items-center">
        {sandboxRun?.created_at ? (
          <Text theme="info" flex className="gap-1">
            <Icon variant="CalendarBlankIcon" />
            <Time variant="subtext" time={sandboxRun.created_at} />
          </Text>
        ) : null}
        {sandboxRun?.created_at ? (
          <Text theme="info" flex className="gap-1">
            <Icon variant="TimerIcon" />
            <Duration variant="subtext" beginTime={sandboxRun.created_at} endTime={sandboxRun.updated_at} />
          </Text>
        ) : null}
        {sandboxRun?.runner_jobs?.at(0)?.install_role_usage?.role_name ? (
          <Text theme="info" flex className="gap-1">
            <Icon variant="FileLockIcon" />
            <Text variant="subtext">{sandboxRun.runner_jobs.at(0).install_role_usage.role_name}</Text>
          </Text>
        ) : null}
      </div>

      {step?.execution_type === 'approval' ? (
        sandboxRun?.log_stream ? (
          <LogStreamProvider
            shouldPoll
            logStreamId={sandboxRun.log_stream.id}
          >
            <UnifiedLogsProvider>
              <LogViewerProvider>
                <Tabs
                  tabs={{
                    plan: (
                      <div className="mt-4">
                        <Plan step={step} />
                      </div>
                    ),
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
            </UnifiedLogsProvider>
          </LogStreamProvider>
        ) : (
          <Tabs
            tabs={{
              plan: (
                <div className="mt-4">
                  <Plan step={step} />
                </div>
              ),
              logs: <SandboxRunLogsSkeleton />,
            }}
          />
        )
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
