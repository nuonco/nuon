import { Badge } from '@/components/common/Badge'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { ChangeCountSummary } from '@/components/approvals/plan-diffs/ChangeCountSummary'
import type { TWorkflowStep } from '@/types'
import { getPolicyViolationCounts, getStepBadges } from '@/utils/workflow-utils'

const humanize = (value: string) => {
  const text = value.replace(/[-_]+/g, ' ')
  return text.charAt(0).toUpperCase() + text.slice(1)
}

export const planStepFor = (
  step: TWorkflowStep,
  steps: TWorkflowStep[]
): TWorkflowStep | undefined => {
  if (step.approval) return step
  if (step.execution_type === 'approval') return undefined

  const component = step.metadata?.component_name
  return steps
    .slice(0, steps.indexOf(step))
    .findLast(
      (candidate) =>
        candidate.approval &&
        candidate.metadata?.component_name === component &&
        candidate.step_target_type === step.step_target_type
    )
}

const Section = ({ children }: { children: React.ReactNode }) => (
  <div className="flex flex-col gap-1 border-t pt-2">{children}</div>
)

export const StepTooltipContent = ({
  step,
  steps,
}: {
  step: TWorkflowStep
  steps: TWorkflowStep[]
}) => {
  const status = step.status?.status ?? 'unknown'
  const badges = getStepBadges(step)
  const lineage = badges.length > 1 ? badges[0] : undefined
  const outcome = badges[badges.length - 1]
  const { denyCount, warnCount, skipFindings } = getPolicyViolationCounts(step)
  const plan = planStepFor(step, steps)?.approval
  const hasPolicy = denyCount > 0 || warnCount > 0 || skipFindings.length > 0

  return (
    <div className="flex w-64 flex-col gap-2 py-1 text-left">
      <div className="flex flex-col gap-1">
        <Status status={status}>
          {String(outcome?.children ?? humanize(status))}
        </Status>
        {step.status?.status_human_description ? (
          <Text variant="subtext" theme="neutral">
            {step.status.status_human_description}
          </Text>
        ) : null}
      </div>

      {lineage ? (
        <Section>
          <Text variant="subtext" theme="neutral">
            Retry
          </Text>
          <Badge {...lineage} size="sm" />
        </Section>
      ) : null}

      {plan ? (
        <Section>
          <Text variant="subtext" theme="neutral">
            Changes
          </Text>
          <ChangeCountSummary
            added={plan.changes_create}
            updated={plan.changes_update}
            removed={plan.changes_delete}
            replaced={plan.changes_replace}
          />
        </Section>
      ) : null}

      {hasPolicy ? (
        <Section>
          <Text variant="subtext" theme="neutral">
            Policies
          </Text>
          {denyCount > 0 ? (
            <Text variant="subtext" weight="strong" theme="error">
              {denyCount} {denyCount === 1 ? 'violation' : 'violations'}
            </Text>
          ) : null}
          {warnCount > 0 ? (
            <Text variant="subtext" weight="strong" theme="warn">
              {warnCount} {warnCount === 1 ? 'warning' : 'warnings'}
            </Text>
          ) : null}
          {skipFindings.length > 0 ? (
            <Text variant="subtext" weight="strong" theme="neutral">
              {skipFindings.length} skipped
            </Text>
          ) : null}
        </Section>
      ) : null}
    </div>
  )
}
