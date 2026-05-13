import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { TWorkflowStep } from '@/types'
import { getPolicyViolationCounts, getStepBanner } from '@/utils/workflow-utils'
import { StepButtons } from './StepButtons'
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
    stepStatus === 'cancelled' ||
    stepStatus === 'discarded'
  const { hasViolations: hasPolicyViolations, hasPolicyData, passedCount } =
    getPolicyViolationCounts(step)

  return (
    <>
      {hasApproval && !planOnly && !isTerminal ? (
        <ApprovalBanner step={step} />
      ) : bannerCfg ? (
        <Banner theme={bannerCfg.theme} onDismiss={onDismiss}>
          <div className="flex items-end justify-between gap-4">
            <div className="flex flex-col">
              <Text weight="strong">{bannerCfg.title}</Text>
              <Text variant="subtext" theme="neutral">
                {bannerCfg.copy}
              </Text>
              {stepStatus === 'error' && statusDescription ? (
                <Text variant="subtext" theme="error">
                  {statusDescription}
                </Text>
              ) : null}
            </div>

            <div className="flex items-end gap-4">
              {onViewDetails ? (
                <Button variant="ghost" size="md" onClick={onViewDetails}>
                  View details <Icon variant="CaretRightIcon" />
                </Button>
              ) : null}
              {bannerCfg.theme === 'error' ? (
                <StepButtons buttonSize="md" step={step} />
              ) : null}
            </div>
          </div>
        </Banner>
      ) : null}
      {hasPolicyViolations ? (
        <PolicyViolations step={step} />
      ) : hasPolicyData && passedCount > 0 ? (
        <Banner theme="success">
          <Text weight="strong">
            All policy checks passed successfully
          </Text>
        </Banner>
      ) : null}
    </>
  )
}
