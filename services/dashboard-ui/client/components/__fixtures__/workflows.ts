import type { TWorkflow, TWorkflowStep } from '@/types'

export const mockWorkflowStep: TWorkflowStep = {
  id: 'step-001',
  install_workflow_id: 'wf-001',
  name: 'deploy-terraform',
  execution_type: 'approval',
  status: {
    status: 'pending',
    status_human_description: 'Waiting for approval',
  },
  approval: {
    id: 'approval-001',
    type: 'terraform_plan',
    response: undefined,
  },
  finished: false,
} as TWorkflowStep

export const mockHelmStep: TWorkflowStep = {
  ...mockWorkflowStep,
  id: 'step-002',
  name: 'deploy-helm',
  approval: {
    id: 'approval-002',
    type: 'helm_approval',
    response: undefined,
  },
} as TWorkflowStep

export const mockK8sStep: TWorkflowStep = {
  ...mockWorkflowStep,
  id: 'step-003',
  name: 'deploy-k8s-manifest',
  approval: {
    id: 'approval-003',
    type: 'kubernetes_manifest_approval',
    response: undefined,
  },
} as TWorkflowStep

export const mockWorkflow: TWorkflow = {
  id: 'wf-001',
  type: 'deploy_components',
  steps: [
    mockWorkflowStep,
    { ...mockWorkflowStep, id: 'step-010', name: 'apply-terraform', execution_type: 'approval', approval: { id: 'a-010', type: 'terraform_plan', response: undefined } },
    { ...mockWorkflowStep, id: 'step-011', name: 'deploy-helm-chart', execution_type: 'approval', approval: { id: 'a-011', type: 'helm_approval', response: undefined } },
  ],
} as TWorkflow
