import { Banner, type TBannerTheme } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import type {
  TWorkflowStep,
  TWorkflowStepApprovalType,
  TWorkflowStepApprovalResponse,
} from '@/types'
import {
  getApprovalType,
  getApprovalResponseType,
} from '@/utils/approval-utils'
import { ApprovePlanButton } from './ApprovePlan'
import { DenyPlanButton } from './DenyPlan'
import { RetryPlanButton } from './RetryPlan'

const APPROVAL_BANNER_COPY: Record<
  Exclude<TWorkflowStepApprovalType, 'approve-all' | 'noop'>,
  { title: string; message: string }
> = {
  terraform_plan: {
    title: 'Terraform plan requires review',
    message:
      'This Terraform plan is ready for your review. Please inspect the proposed infrastructure changes before applying them.',
  },
  helm_approval: {
    title: 'Helm chart plan requires review',
    message:
      'This Helm chart plan includes updates to your deployment. Review the changes to ensure your release will work as intended.',
  },
  kubernetes_manifest_approval: {
    title: 'Kubernetes manifest requires review',
    message:
      'This Kubernetes manifest contains pending configuration changes. Please review these updates before applying them to your cluster.',
  },
}

interface IApprovalBanner {
  step: TWorkflowStep
}

export const ApprovalBanner = ({ step }: IApprovalBanner) => {
  return step?.approval?.response ? (
    <ApprovedPlanBanner step={step} />
  ) : (
    <AwaitingApprovalBanner step={step} />
  )
}

const AwaitingApprovalBanner = ({ step }: IApprovalBanner) => {
  const bannerCopy = APPROVAL_BANNER_COPY[step?.approval?.type]

  return (
    <Banner className="@container" theme="warn">
      <div className="flex flex-col gap-2">
        <div className="flex flex-col">
          <Text weight="strong">{bannerCopy.title}</Text>
          <Text variant="subtext" theme="neutral">
            {bannerCopy.message}
          </Text>
        </div>

        <div className="flex self-end gap-2">
          <RetryPlanButton step={step} />
          <DenyPlanButton step={step} />
          <ApprovePlanButton step={step} variant="primary" />
        </div>
      </div>
    </Banner>
  )  
}

const RESPONSE_THEME: Record<
  TWorkflowStepApprovalResponse['type'],
  TBannerTheme
> = {
  approve: 'success',
  'auto-approve': 'success',
  deny: 'warn',
  skip: 'default',
  retry: 'info',
}

const ApprovedPlanBanner = ({ step }: IApprovalBanner) => {
  const responseType = getApprovalResponseType(step?.approval?.response?.type)

  return (
    <Banner theme={RESPONSE_THEME[step?.approval?.response?.type]}>
      <div className="flex items-center justify-between gap-4">
        <div className="flex flex-col">
          <Text weight="strong">Plan was {responseType}</Text>
          <Text variant="subtext" theme="neutral">
            This {getApprovalType(step?.approval?.type)} plan was {responseType}
            .
          </Text>
        </div>
      </div>
    </Banner>
  )
}
