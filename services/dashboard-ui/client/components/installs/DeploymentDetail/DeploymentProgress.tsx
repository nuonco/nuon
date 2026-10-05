import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TWorkflowStep } from '@/types'
import { humanize } from '@/utils/string-utils'
import {
  deploymentStepContext,
  isAwaitingDeploymentApproval,
  isDeploymentRunning,
  type TDeploymentOutcome,
} from './deployment-progress'

export type TDeploymentRun = {
  status: string
  activity: string
  steps: TWorkflowStep[]
  outcomes: TDeploymentOutcome[]
}

export const DeploymentRunStatus = ({ run }: { run: TDeploymentRun }) => (
  <span className="flex flex-wrap items-center gap-2">
    <Status status={run.status} variant="badge" />
    {run.steps.some(isAwaitingDeploymentApproval) ? (
      <Badge size="sm" theme="warn">
        Pending approval
      </Badge>
    ) : null}
  </span>
)

export const ResourceScopeSummary = ({ run }: { run: TDeploymentRun }) => (
  <div
    role="group"
    aria-label="Resource outcomes"
    className="flex flex-wrap items-center gap-x-5 gap-y-2"
  >
    {[...new Set(run.outcomes.map((outcome) => outcome.category))].map(
      (category) => {
        const outcomes = run.outcomes.filter(
          (outcome) => outcome.category === category
        )
        const completed = outcomes.every(
          (outcome) => outcome.status === 'success'
        )
        const partial =
          !completed && outcomes.some((outcome) => outcome.status === 'success')
        const status = completed
          ? 'success'
          : outcomes.some((outcome) => outcome.status === 'in-progress')
            ? 'in-progress'
            : partial
              ? 'warn'
              : outcomes.some((outcome) => outcome.status === 'error')
                ? 'error'
                : outcomes.every((outcome) => outcome.status === 'not-started')
                  ? 'not-started'
                  : 'unknown'
        const label = completed
          ? 'Completed'
          : status === 'in-progress'
            ? 'In progress'
            : partial
              ? 'Partial rollout'
              : status === 'error'
                ? 'Failed'
                : status === 'not-started'
                  ? 'Not started'
                  : 'Outcome unknown'
        return (
          <span
            key={category}
            role="group"
            aria-label={`${category}: ${label}`}
            className="flex items-center gap-2"
          >
            {partial && status !== 'in-progress' ? (
              <Icon variant="CircleHalfIcon" theme="warn" />
            ) : (
              <Status
                status={status}
                variant="timeline"
                isWithoutText
                iconSize={14}
              />
            )}
            <Text>
              {category}
              {!completed ? ` — ${label}` : ''}
            </Text>
          </span>
        )
      }
    )}
  </div>
)

export const ResourceOutcomes = ({ run }: { run: TDeploymentRun }) =>
  run.outcomes.length ? (
    <section
      aria-label="Resource rollout outcomes"
      className="flex flex-col gap-3"
    >
      <Text weight="strong">Resource outcomes</Text>
      <ul className="flex flex-col gap-2">
        {run.outcomes.map((outcome) => (
          <li
            key={`${outcome.category}-${outcome.name}`}
            aria-label={`${outcome.name}: ${outcome.detail}`}
            className="flex items-start gap-3"
          >
            <Status
              status={outcome.status}
              variant="timeline"
              isWithoutText
              iconSize={14}
            />
            <div className="flex min-w-0 flex-wrap items-baseline gap-x-3">
              <Text
                family={
                  outcome.category === 'Components' ||
                  outcome.category === 'Images'
                    ? 'mono'
                    : 'sans'
                }
                weight="strong"
              >
                {outcome.name}
              </Text>
              <Text theme={outcome.status === 'error' ? 'error' : 'neutral'}>
                {outcome.detail}
              </Text>
            </div>
          </li>
        ))}
      </ul>
    </section>
  ) : null

export const StepContext = ({ run }: { run: TDeploymentRun }) => {
  const { current, next } = deploymentStepContext(run.status, run.steps)
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
              : 'Step unavailable')}
        </Text>
        {current && !isDeploymentRunning(run.status) ? (
          <Text variant="subtext" theme="neutral">
            {run.status === 'success' ? 'Completed' : 'Failed'}
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
            (run.steps.length ? 'No remaining steps' : 'Step unavailable')}
        </Text>
        {next && run.status === 'error' ? (
          <Text variant="subtext" theme="neutral">
            Not started — deployment stopped
          </Text>
        ) : null}
      </div>
    </div>
  )
}

export const StepProgress = ({ run }: { run: TDeploymentRun }) => {
  if (!run.steps.length || !isDeploymentRunning(run.status)) return null
  const completed = run.steps.filter(
    (step) => step.status?.status === 'success'
  ).length
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
            className={`h-1 flex-1 rounded-sm ${step.status?.status === 'success' ? 'bg-green-500' : 'bg-cool-grey-200 dark:bg-dark-grey-600'}`}
          />
        ))}
      </div>
      <Text variant="subtext" theme="neutral">
        {completed}/{run.steps.length} steps complete
      </Text>
    </div>
  )
}
