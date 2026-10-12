import { Badge } from '@/components/common/Badge'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { humanize } from '@/utils/string-utils'
import {
  activeStepDetail,
  deploymentStepContext,
  isAwaitingDeploymentApproval,
  isAwaitingDeploymentRetry,
  isCompletedDeploymentStep,
  isDeploymentRunning,
  isQueuedDeployment,
  type TDeploymentOutcome,
  type TDeploymentStep,
} from './deployment-progress'

export type TDeploymentRun = {
  status: string
  activity: string
  steps: TDeploymentStep[]
  outcomes: TDeploymentOutcome[]
}

export const DeploymentRunStatus = ({
  run,
}: {
  run: Pick<TDeploymentRun, 'status' | 'steps'>
}) => (
  <span className="flex flex-wrap items-center gap-2">
    <Status status={run.status} variant="badge" />
    {run.steps.some(isAwaitingDeploymentApproval) ? (
      <Badge size="sm" theme="warn">
        Pending approval
      </Badge>
    ) : null}
  </span>
)

export const StepContext = ({ run }: { run: TDeploymentRun }) => {
  const { current, next } = deploymentStepContext(run.status, run.steps)
  const preparing = !run.steps.length && isQueuedDeployment(run.status)
  const activeDetail = activeStepDetail(current)
  return (
    <div className="grid grid-cols-1 gap-4 @sm:grid-cols-2">
      <div
        role="group"
        aria-label="Current step"
        className="flex min-w-0 flex-col gap-1"
      >
        <Text variant="body" theme="neutral">
          Current step
        </Text>
        <Text
          variant="base"
          weight="strong"
          theme={run.status === 'error' ? 'error' : 'default'}
          className="break-words"
        >
          {current?.name ??
            (isDeploymentRunning(run.status)
              ? run.activity
              : preparing
                ? 'Preparing steps'
                : 'Step unavailable')}
        </Text>
        {current && isDeploymentRunning(run.status) && activeDetail ? (
          <Text variant="subtext" theme="neutral">
            {activeDetail.charAt(0).toUpperCase() + activeDetail.slice(1)}
          </Text>
        ) : null}
        {current && !isDeploymentRunning(run.status) ? (
          <Text variant="subtext" theme="neutral">
            {run.status === 'success'
              ? 'Completed'
              : run.status === 'cancelled'
                ? 'Cancelled'
                : 'Failed'}
          </Text>
        ) : null}
      </div>
      <div
        role="group"
        aria-label="Next step"
        className="flex min-w-0 flex-col gap-1 @sm:border-l @sm:pl-4"
      >
        <Text variant="body" theme="neutral">
          Up next
        </Text>
        <Text variant="base" theme="neutral" className="break-words">
          {next?.name ??
            (run.steps.length
              ? 'No remaining steps'
              : preparing
                ? 'Waiting for steps'
                : 'Step unavailable')}
        </Text>
        {next && ['error', 'cancelled'].includes(run.status) ? (
          <Text variant="subtext" theme="neutral">
            Not started — deployment stopped
          </Text>
        ) : null}
      </div>
    </div>
  )
}

export const StepProgress = ({ run }: { run: TDeploymentRun }) => {
  if (
    !run.steps.length ||
    (!isDeploymentRunning(run.status) && !isAwaitingDeploymentRetry(run.status))
  )
    return null
  const completed = run.steps.filter(isCompletedDeploymentStep).length
  const description = run.steps
    .map((step) => `${step.name}: ${humanize(step.status?.status)}`)
    .join('; ')
  return (
    <div
      className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3"
      role="group"
      aria-label={`Workflow steps: ${completed} of ${run.steps.length} complete. ${description}`}
      title={description}
    >
      <div className="flex min-w-0 gap-1" aria-hidden="true">
        {run.steps.map((step) => (
          <span
            key={step.id}
            className={`h-1 flex-1 rounded-sm ${isCompletedDeploymentStep(step) ? 'bg-green-500' : 'bg-cool-grey-200 dark:bg-dark-grey-600'}`}
          />
        ))}
      </div>
      <Text variant="subtext" theme="neutral">
        {completed}/{run.steps.length} steps complete
      </Text>
    </div>
  )
}
