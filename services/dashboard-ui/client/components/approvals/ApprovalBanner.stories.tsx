import { ApprovalBanner } from './ApprovalBanner'
import type { TWorkflowStep } from '@/types'

export default { title: 'Approvals/ApprovalBanner' }

const baseTerraformStep: TWorkflowStep = {
  id: 'step-1',
  approval: {
    type: 'terraform_plan',
    response: undefined,
  },
  status: {
    status: 'pending',
  },
} as TWorkflowStep

const approvedStep: TWorkflowStep = {
  ...baseTerraformStep,
  approval: {
    type: 'terraform_plan',
    response: { type: 'approve' },
  },
} as TWorkflowStep

const deniedStep: TWorkflowStep = {
  ...baseTerraformStep,
  approval: {
    type: 'terraform_plan',
    response: { type: 'deny' },
  },
} as TWorkflowStep

const helmStep: TWorkflowStep = {
  ...baseTerraformStep,
  approval: {
    type: 'helm_approval',
    response: undefined,
  },
} as TWorkflowStep

const k8sStep: TWorkflowStep = {
  ...baseTerraformStep,
  approval: {
    type: 'kubernetes_manifest_approval',
    response: undefined,
  },
} as TWorkflowStep

const supersededStep: TWorkflowStep = {
  ...baseTerraformStep,
  execution_type: 'approval',
  retried: true,
  approval: {
    type: 'helm_approval',
    response: { type: 'approve' },
  },
  status: {
    status: 'success',
    metadata: { status: 'approved' },
  },
} as TWorkflowStep

export const AwaitingTerraformApproval = () => <ApprovalBanner step={baseTerraformStep} />

export const Approved = () => <ApprovalBanner step={approvedStep} />

export const Superseded = () => <ApprovalBanner step={supersededStep} />

export const Denied = () => <ApprovalBanner step={deniedStep} />

export const AwaitingHelmApproval = () => <ApprovalBanner step={helmStep} />

export const AwaitingKubernetesApproval = () => <ApprovalBanner step={k8sStep} />
