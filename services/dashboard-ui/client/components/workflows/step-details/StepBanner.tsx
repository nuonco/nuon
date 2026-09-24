import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { CompositeError } from '@/components/common/CompositeError'
import { Text } from '@/components/common/Text'
import { useStepErrorLog } from '@/hooks/use-step-error-log'
import type { TWorkflowStep } from '@/types'
import { getPolicyViolationCounts, getStepBanner } from '@/utils/workflow-utils'
import { StepButtons } from './StepButtons'
import { StepErrorLog } from './StepErrorLog'
import { PolicyViolations } from './PolicyViolations'

export const StepBanner = ({
  step,
  planOnly = false,
  onDismiss,
  onViewDetails,
}: {
  step: TWorkflowStep
  planOnly?: boolean
  onDismiss?: () => void
  onViewDetails?: () => void
}) => {
  const hasApproval = Boolean(step?.approval)
  const bannerCfg = getStepBanner(step)
  const stepStatus = step?.status?.status
  const statusDescription = step?.status?.status_human_description?.replace(
    /\s*\(type:\s*\w+,\s*retryable:\s*\w+\)\s*$/,
    ''
  )
  const isTerminal =
    stepStatus === 'error' ||
    stepStatus === 'failed-pending-retry' ||
    stepStatus === 'cancelled' ||
    stepStatus === 'discarded'
  const {
    hasViolations: hasPolicyViolations,
    hasPolicyData,
    passedCount,
  } = getPolicyViolationCounts(step)
  const metadata = step?.status?.metadata as Record<string, unknown> | undefined
  const isPolicyAutoApproved =
    metadata?.auto_approved && metadata?.check === 'policy-auto-approve'
  const compositeError = step?.status?.composite_error
  const showCompositeError =
    Boolean(compositeError) && bannerCfg?.theme === 'error'
  const errorLog = useStepErrorLog(step, {
    enabled: !showCompositeError && isTerminal && bannerCfg?.theme === 'error',
  })

  return (
    <div className="flex flex-col gap-2">
      {hasApproval && !planOnly && !isTerminal ? (
        <ApprovalBanner step={step} />
      ) : bannerCfg ? (
        <div className="flex flex-col gap-2">
          <Banner theme={bannerCfg.theme} onDismiss={onDismiss}>
            <div className="flex items-end justify-between gap-4">
              <div className="flex flex-col min-w-0">
                <Text weight="strong" className="break-words">
                  {bannerCfg.title}
                </Text>
                <Text variant="subtext" theme="neutral" className="break-words">
                  {bannerCfg.copy}
                </Text>
                {(stepStatus === 'error' ||
                  stepStatus === 'failed-pending-retry') &&
                statusDescription &&
                !showCompositeError &&
                bannerCfg.theme === 'error' ? (
                  <Text
                    variant="subtext"
                    theme="error"
                    className="break-words"
                  >
                    {statusDescription}
                  </Text>
                ) : null}
              </div>

              <div className="flex items-end gap-4">
                {onViewDetails ? (
                  <Button variant="ghost" size="md" onClick={onViewDetails}>
                    View details
                  </Button>
                ) : null}
                {bannerCfg.theme === 'error' ? (
                  <StepButtons buttonSize="md" step={step} />
                ) : null}
              </div>
            </div>
          </Banner>
          {showCompositeError && compositeError ? (
            <CompositeError error={compositeError} />
          ) : errorLog ? (
            <StepErrorLog text={errorLog.detail ?? errorLog.message} />
          ) : null}
        </div>
      ) : null}
      {hasPolicyViolations ? (
        <PolicyViolations step={step} />
      ) : hasPolicyData && passedCount > 0 ? (
        <Banner theme="success">
          <Text weight="strong">
            {isPolicyAutoApproved
              ? 'Auto-approved: all policy checks passed'
              : 'All policy checks passed'}
          </Text>
        </Banner>
      ) : null}
    </div>
  )
}
