import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import type { TWorkflowStep } from '@/types'
import { getPolicyViolationCounts, getStepBanner } from '@/utils/workflow-utils'
import { StepButtons } from './StepButtons'
import { PolicyViolations } from './PolicyViolations'

export const StepBanner = ({
  step,
  planOnly = false,
}: {
  step: TWorkflowStep
  planOnly?: boolean
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
        <Banner theme={bannerCfg.theme}>
          <div className="flex items-center justify-between gap-4">
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

            <div className="flex gap-4">
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
