import {
  TWorkflowStepApproval,
  TWorkflowStepApprovalType,
  TWorkflowStep,
} from '@/types/ctl-api.types'

export const mockWorkflowStepApproval: TWorkflowStepApproval = {
  id: 'approval-123e4567-e89b-12d3-a456-426614174000',
  created_at: '2024-01-15T10:30:00.000Z',
  updated_at: '2024-01-15T10:35:00.000Z',
  created_by_id: 'user-123e4567-e89b-12d3-a456-426614174001',
  installWorkflowStepID: 'step-123e4567-e89b-12d3-a456-426614174002',
  workflow_step_id: 'step-123e4567-e89b-12d3-a456-426614174002',
  runner_job_id: 'job-123e4567-e89b-12d3-a456-426614174003',
  owner_id: 'org-123e4567-e89b-12d3-a456-426614174004',
  owner_type: 'organization',
  type: 'manual' as TWorkflowStepApprovalType,
  response: undefined, // {
  //   id: 'response-123e4567-e89b-12d3-a456-426614174005',
  //   type: 'deny',
  //   created_at: '2024-01-15T10:35:00.000Z',
  //   created_by_id: 'user-123e4567-e89b-12d3-a456-426614174001',
  //   note: 'Deployment looks good, approving to proceed',
  // }
}

export const createMockWorkflowStepApproval = (
  overrides?: Partial<TWorkflowStepApproval>
): TWorkflowStepApproval => ({
  ...mockWorkflowStepApproval,
  ...overrides,
})

export const mockWorkflowStep: TWorkflowStep = {
  id: 'step-123e4567-e89b-12d3-a456-426614174000',
  created_at: '2024-01-15T10:00:00.000Z',
  updated_at: '2024-01-15T10:30:00.000Z',
  started_at: '2024-01-15T10:05:00.000Z',
  finished_at: '2024-01-15T10:30:00.000Z',
  name: 'Deploy Infrastructure',
  idx: 1,
  group_idx: 0,
  group_retry_idx: 0,
  workflow_id: 'workflow-123e4567-e89b-12d3-a456-426614174001',
  install_workflow_id: 'workflow-123e4567-e89b-12d3-a456-426614174001',
  owner_id: 'org-123e4567-e89b-12d3-a456-426614174002',
  owner_type: 'organization',
  created_by_id: 'user-123e4567-e89b-12d3-a456-426614174003',
  execution_type: 'system',
  execution_time: 1500,
  finished: true,
  retried: false,
  retryable: true,
  skippable: false,
  step_target_id: 'deploy-123e4567-e89b-12d3-a456-426614174004',
  step_target_type: 'install_deploy',
  status: {
    status: 'cancelled',
    status_human_description: 'tktk',
  },
  metadata: {
    'terraform.plan': 'true',
    'component.name': 'web-app',
    'deployment.environment': 'production',
  },
  links: {
    logs: '/api/logs/step-123e4567-e89b-12d3-a456-426614174000',
    artifacts: '/api/artifacts/step-123e4567-e89b-12d3-a456-426614174000',
  },
}

export const createMockWorkflowStep = (
  overrides?: Partial<TWorkflowStep>
): TWorkflowStep => ({
  ...mockWorkflowStep,
  ...overrides,
})
