import { Plan } from '@/components/approvals/Plan'
import { CompositeError } from '@/components/common/CompositeError'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { SSELogs } from '@/components/log-stream/SSELogs'
import { TraceView } from '@/components/spans/TraceView'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import type { TDeploy, TWorkflowStep } from '@/types'
import { DeployApply } from '../DeployApply'

export interface IDeployStepDetails {
  step?: TWorkflowStep
  orgId: string
  deploy?: TDeploy
  error: any
  isLoading: boolean
}

export const DeployStepDetails = ({
  step,
  orgId,
  deploy,
  error,
  isLoading,
}: IDeployStepDetails) => {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-4">
        {isLoading && !deploy ? (
          <DeployStepDetailsSkeleton />
        ) : error ? (
          <Text variant="base" weight="strong" theme="error">
            Unable to load deploy details
          </Text>
        ) : (
          <>
            <Text variant="base" weight="strong">
              {deploy?.component_name} deployment
            </Text>
            {deploy?.component_id ? (
              <Text variant="subtext">
                <Link
                  href={`/${orgId}/installs/${step?.owner_id}/components/${deploy.component_id}`}
                >
                  View component <Icon variant="CaretRightIcon" />
                </Link>
              </Text>
            ) : null}

            {deploy?.component_id && deploy?.id ? (
              <Text variant="subtext">
                <Link
                  href={`/${orgId}/installs/${step?.owner_id}/components/${deploy.component_id}/deploys/${deploy.id}`}
                >
                  View deploy logs <Icon variant="CaretRightIcon" />
                </Link>
              </Text>
            ) : null}
          </>
        )}
      </div>
      <div className="flex flex-wrap gap-x-4 gap-y-1 items-center">
        {deploy?.created_at ? (
          <Text theme="info" flex className="gap-1">
            <Icon variant="CalendarBlankIcon" />
            <Time variant="subtext" time={deploy.created_at} />
          </Text>
        ) : null}
        {deploy?.created_at ? (
          <Text theme="info" flex className="gap-1">
            <Icon variant="TimerIcon" />
            <Duration variant="subtext" beginTime={deploy.created_at} endTime={deploy.updated_at} />
          </Text>
        ) : null}
        {deploy?.runner_jobs?.at(0)?.install_role_usage?.role_name ? (
          <Text theme="info" flex className="gap-1">
            <Icon variant="FileLockIcon" />
            <Text variant="subtext">{deploy?.runner_jobs?.at(0)?.install_role_usage?.role_name}</Text>
          </Text>
        ) : null}
      </div>
      {deploy?.composite_error ? (
        <CompositeError error={deploy.composite_error} />
      ) : null}
      {step?.execution_type === 'approval' ? (
        <ApprovalStepTabs step={step} deploy={deploy} />
      ) : (
        <DeployApply initDeploy={deploy} />
      )}
    </div>
  )
}

// Plan tab is only rendered once the runner has produced an approval
// (step.approval set). When present it's the first tab so finished
// approval steps land on Plan; otherwise Logs is first.
const ApprovalStepTabs = ({
  step,
  deploy,
}: {
  step: TWorkflowStep
  deploy?: TDeploy
}) => {
  const hasPlan = !!step?.approval
  const hasLogStream = !!deploy?.log_stream

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
      logStreamId={deploy!.log_stream!.id}
      shouldPoll={deploy!.log_stream!.open}
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
    <LogStreamProvider logStreamId={deploy!.log_stream!.id}>
      <LogViewerProvider>{tabsEl}</LogViewerProvider>
    </LogStreamProvider>
  )
}

export const DeployStepDetailsSkeleton = () => {
  return (
    <>
      <Skeleton height="24px" width="180px" />
      <Skeleton height="17px" width="115px" />
      <Skeleton height="17px" width="115px" />
    </>
  )
}
