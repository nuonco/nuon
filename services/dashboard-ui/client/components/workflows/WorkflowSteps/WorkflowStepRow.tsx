import { Badge } from '@/components/common/Badge'
import { Duration } from '@/components/common/Duration'
import { Text } from '@/components/common/Text'
import { PolicyCountsBadge } from '@/components/workflows/step-details/PolicyCountsBadge'
import { StepButtons } from '@/components/workflows/step-details/StepButtons'
import { StepDetailPanelButton } from '@/components/workflows/step-details/StepDetailPanel'
import { StepTitle } from '@/components/workflows/step-details/StepTitle'
import type { TWorkflowStep } from '@/types'
import { cn } from '@/utils/classnames'
import { getStepBadges } from '@/utils/workflow-utils'

export interface IWorkflowStepRow {
  step: TWorkflowStep
  approvalPrompt?: boolean
  planOnly?: boolean
  showRetry?: boolean
  nested?: boolean
  attemptNumber?: number
}

export const WorkflowStepRow = ({
  step,
  approvalPrompt = false,
  planOnly = false,
  showRetry = false,
  nested = false,
  attemptNumber,
}: IWorkflowStepRow) => {
  const badges = getStepBadges(step, approvalPrompt, planOnly)

  const hideDetails =
    (step?.execution_type === 'system' && !step?.step_target_type) ||
    step?.status?.status === 'pending' ||
    step?.status?.status === 'not-attempted'

  return (
    <div
      className={cn(
        'flex flex-col md:flex-row md:items-center gap-4 w-full',
        nested && 'opacity-80'
      )}
    >
      <StepTitle step={step} />

      <div className="flex items-center flex-wrap gap-2 md:gap-4">
        {attemptNumber ? (
          <Badge theme="neutral" size="sm">
            Attempt {attemptNumber}
          </Badge>
        ) : null}

        {badges.map((badge) => (
          <Badge key={badge.children} {...badge} size="sm" />
        ))}

        <PolicyCountsBadge step={step} />

        {hideDetails ? null : (
          <StepDetailPanelButton
            approvalPrompt={approvalPrompt}
            step={step}
            planOnly={planOnly}
          />
        )}

        {step?.finished ? (
          <Text variant="subtext" theme="neutral">
            Completed in{' '}
            <Duration variant="subtext" nanoseconds={step?.execution_time} />
          </Text>
        ) : null}
      </div>

      <StepButtons
        isApproveAll={!approvalPrompt}
        showRetry={showRetry}
        step={step}
      />
    </div>
  )
}
