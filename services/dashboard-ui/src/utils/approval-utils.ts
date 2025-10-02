import type {
  TWorkflowStepApprovalResponse,
  TWorkflowStepApprovalType,
} from '@/types'

const APPROVAL_TYPE: Record<TWorkflowStepApprovalType, string> = {
  'approve-all': 'all changes approved',
  terraform_plan: 'terraform',
  kubernetes_manifest_approval: 'kubernetes',
  helm_approval: 'helm',
  noop: 'no-op',
}

export function getApprovalType(
  approvalType: TWorkflowStepApprovalType
): string {
  return APPROVAL_TYPE[approvalType]
}

const RESPONSE_TYPE: Record<TWorkflowStepApprovalResponse['type'], string> = {
  approve: 'approved',
  'auto-approve': 'auto-approved',
  deny: 'denied',
  retry: 'retired',
  skip: 'skipped',
}

export function getApprovalResponseType(
  responseType: TWorkflowStepApprovalResponse['type']
): string {
  return RESPONSE_TYPE[responseType]
}

export const APPROVAL_MODAL_COPY: Record<
  Exclude<TWorkflowStepApprovalType, 'approve-all' | 'noop'>,
  { title: string; heading: string; message: string }
> = {
  terraform_plan: {
    title: 'Approve Terraform plan?',
    heading: 'Are you sure you want to approve these infrastructure changes?',
    message:
      'Approving the plan will immediately apply the proposed changes to your environment.',
  },
  helm_approval: {
    title: 'Approve Helm chart plan?',
    heading: 'Are you sure you want to approve these deployment changes?',
    message:
      'Approving the plan will immediately apply the proposed updates to your release.',
  },
  kubernetes_manifest_approval: {
    title: 'Approve Kubernetes manifest?',
    heading: 'Are you sure you want to approve these configuration changes?',
    message:
      'Approving the manifest will immediately apply the changes to your cluster.',
  },
}

export const DENY_MODAL_COPY: Record<
  Exclude<TWorkflowStepApprovalType, 'approve-all' | 'noop'>,
  { title: string; heading: string; message: string }
> = {
  terraform_plan: {
    title: 'Deny Terraform plan?',
    heading: 'Are you sure you want to deny these infrastructure changes?',
    message:
      'Denying the plan will discard the proposed changes and prevent them from being applied.',
  },
  helm_approval: {
    title: 'Deny Helm chart plan?',
    heading: 'Are you sure you want to deny these deployment changes?',
    message:
      'Denying the plan will discard the proposed updates and prevent them from being applied.',
  },
  kubernetes_manifest_approval: {
    title: 'Deny Kubernetes manifest?',
    heading: 'Are you sure you want to deny these configuration changes?',
    message:
      'Denying the manifest will discard the changes and prevent them from being applied to your cluster.',
  },
}

export const RETRY_MODAL_COPY: Record<
  Exclude<TWorkflowStepApprovalType, 'approve-all' | 'noop'>,
  { title: string; heading: string; message: string }
> = {
  terraform_plan: {
    title: 'Retry Terraform plan?',
    heading: 'Are you sure you want to retry this Terraform plan?',
    message:
      'Retrying will generate a new plan, replacing the current proposed infrastructure changes.',
  },
  helm_approval: {
    title: 'Retry Helm chart plan?',
    heading: 'Are you sure you want to retry this Helm chart plan?',
    message:
      'Retrying will generate a new plan, replacing the current proposed deployment updates.',
  },
  kubernetes_manifest_approval: {
    title: 'Retry Kubernetes manifest?',
    heading: 'Are you sure you want to retry this Kubernetes manifest?',
    message:
      'Retrying will generate a new manifest, replacing the current proposed configuration changes for your cluster.',
  },
}
