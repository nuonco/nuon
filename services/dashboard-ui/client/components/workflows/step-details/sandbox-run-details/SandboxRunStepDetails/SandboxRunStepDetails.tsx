import { Plan } from '@/components/approvals/Plan'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { SSELogs } from '@/components/log-stream/SSELogs'
import { TraceView } from '@/components/spans/TraceView'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
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

        {step?.owner_id ? (
          <Text variant="subtext">
            <Link href={`/${orgId}/installs/${step.owner_id}/sandbox`}>
              View sandbox <Icon variant="CaretRightIcon" />
            </Link>
          </Text>
        ) : null}

        {step?.owner_id && step?.step_target_id ? (
          <Text variant="subtext">
            <Link
              href={`/${orgId}/installs/${step.owner_id}/sandbox/runs/${step.step_target_id}`}
            >
              View run logs <Icon variant="CaretRightIcon" />
            </Link>
          </Text>
        ) : null}
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
        <ApprovalStepTabs step={step} sandboxRun={sandboxRun} />
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

// Plan tab is only rendered once the runner has produced an approval
// (step.approval set). When present it's the first tab so finished
// approval steps land on Plan; otherwise Logs is first.
const ApprovalStepTabs = ({
  step,
  sandboxRun,
}: {
  step: TWorkflowStep
  sandboxRun?: TSandboxRun
}) => {
  const hasPlan = !!step?.approval
  const hasLogStream = !!sandboxRun?.log_stream

  const logsTab = hasLogStream ? (
    <SSELogs />
  ) : (
    <EmptyState
      variant="history"
      emptyTitle="Waiting for logs"
      emptyMessage="Logs will appear here as soon as the runner starts streaming them."
    />
  )

  const traceTab = hasLogStream ? (
    <TraceView
      logStreamId={sandboxRun!.log_stream!.id}
      shouldPoll={sandboxRun!.log_stream!.open}
    />
  ) : null

  const tabs: Record<string, React.ReactNode> = hasPlan
    ? {
        plan: (
          <div className="mt-4">
            <Plan step={step} />
          </div>
        ),
        logs: logsTab,
        ...(traceTab ? { trace: traceTab } : {}),
      }
    : {
        logs: logsTab,
        ...(traceTab ? { trace: traceTab } : {}),
      }

  const tabsEl = <Tabs tabs={tabs} />

  if (!hasLogStream) return tabsEl

  return (
    <LogStreamProvider logStreamId={sandboxRun!.log_stream!.id}>
      <LogViewerProvider>{tabsEl}</LogViewerProvider>
    </LogStreamProvider>
  )
}
