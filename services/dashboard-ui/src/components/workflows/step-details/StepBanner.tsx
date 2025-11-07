import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import type { TWorkflowStep } from '@/types'
import { getStepBanner } from '@/utils/workflow-utils'
import { StepButtons } from './StepButtons'

export const StepBanner = ({ step, planOnly = false }: { step: TWorkflowStep, planOnly?: boolean }) => {
  const hasApproval = Boolean(step?.approval)
  const bannerCfg = getStepBanner(step)

  return hasApproval && !planOnly ? (
    <ApprovalBanner step={step} />
  ) : bannerCfg ? (
    <Banner theme={bannerCfg.theme}>
      <div className="flex items-center justify-between gap-4">
        <div className="flex flex-col">
          <Text weight="strong">{bannerCfg.title}</Text>
          <Text variant="subtext" theme="neutral">
            {bannerCfg.copy}
          </Text>
        </div>

        <div className="flex gap-4">
          {bannerCfg.theme === 'error' ? (
            <StepButtons buttonSize="md" step={step} />
          ) : null}
        </div>
      </div>
    </Banner>
  ) : null
}
