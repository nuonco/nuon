import type {
  TWorkflowStepApprovalResponse,
  TWorkflowStepApprovalType,
} from '@/types'

const APPROVAL_TYPE: Record<TWorkflowStepApprovalType, string> = {
  'approve-all': 'all changes approved',
  terraform_plan: 'terraform',
  kubernetes_manifest_approval: 'kubernetes',
  helm_approval: 'helm',
  pulumi_plan: 'pulumi',
  app_branch_plan: 'install group plan',
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
    heading: 'This will immediately apply the proposed infrastructure changes.',
    message:
      'The Terraform plan will be applied to your environment.',
  },
  helm_approval: {
    title: 'Approve Helm chart plan?',
    heading: 'This will immediately apply the proposed deployment changes.',
    message:
      'The Helm chart updates will be applied to your release.',
  },
  kubernetes_manifest_approval: {
    title: 'Approve Kubernetes manifest?',
    heading: 'This will immediately apply the proposed configuration changes.',
    message:
      'The manifest changes will be applied to your cluster.',
  },
  pulumi_plan: {
    title: 'Approve Pulumi plan?',
    heading: 'This will immediately apply the proposed infrastructure changes.',
    message:
      'The Pulumi plan will be applied to your environment.',
  },
  app_branch_plan: {
    title: 'Approve install group plan?',
    heading: 'This will deploy the planned changes to the install group.',
    message:
      'The install group plan will be applied.',
  },
}

export const DENY_MODAL_COPY: Record<
  Exclude<TWorkflowStepApprovalType, 'approve-all' | 'noop'>,
  { title: string; heading: string; message: string }
> = {
  terraform_plan: {
    title: 'Deny Terraform plan?',
    heading: 'The proposed infrastructure changes will be discarded.',
    message:
      'Denying prevents the plan from being applied to your environment.',
  },
  helm_approval: {
    title: 'Deny Helm chart plan?',
    heading: 'The proposed deployment changes will be discarded.',
    message:
      'Denying prevents the updates from being applied to your release.',
  },
  kubernetes_manifest_approval: {
    title: 'Deny Kubernetes manifest?',
    heading: 'The proposed configuration changes will be discarded.',
    message:
      'Denying prevents the changes from being applied to your cluster.',
  },
  pulumi_plan: {
    title: 'Deny Pulumi plan?',
    heading: 'The proposed infrastructure changes will be discarded.',
    message:
      'Denying prevents the Pulumi plan from being applied to your environment.',
  },
  app_branch_plan: {
    title: 'Skip install group plan?',
    heading: 'The planned changes will be skipped.',
    message:
      'Skipping prevents the plan from being deployed to the install group.',
  },
}

export const RETRY_MODAL_COPY: Record<
  Exclude<TWorkflowStepApprovalType, 'approve-all' | 'noop'>,
  { title: string; heading: string; message: string }
> = {
  terraform_plan: {
    title: 'Retry Terraform plan?',
    heading: 'A new plan will be generated, replacing the current proposed changes.',
    message:
      'The existing infrastructure changes will be discarded.',
  },
  helm_approval: {
    title: 'Retry Helm chart plan?',
    heading: 'A new plan will be generated, replacing the current proposed changes.',
    message:
      'The existing deployment updates will be discarded.',
  },
  kubernetes_manifest_approval: {
    title: 'Retry Kubernetes manifest?',
    heading: 'A new manifest will be generated, replacing the current proposed changes.',
    message:
      'The existing configuration changes will be discarded.',
  },
  pulumi_plan: {
    title: 'Retry Pulumi plan?',
    heading: 'A new plan will be generated, replacing the current proposed changes.',
    message:
      'The existing Pulumi infrastructure changes will be discarded.',
  },
  app_branch_plan: {
    title: 'Retry install group plan?',
    heading: 'A new plan will be generated for the install group.',
    message:
      'The existing plan will be discarded.',
  },
}
