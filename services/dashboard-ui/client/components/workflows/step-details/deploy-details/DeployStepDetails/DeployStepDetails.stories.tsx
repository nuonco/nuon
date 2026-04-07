export default {
  title: 'Workflows/DeployStepDetails',
}

import { DeployStepDetails, DeployStepDetailsSkeleton } from './DeployStepDetails'
import type { TWorkflowStep, TDeploy } from '@/types'

const mockStep = {
  id: 'step-1',
  name: 'deploy component',
  step_target_type: 'install_deploys',
  step_target_id: 'deploy-1',
  owner_id: 'inst-1',
  execution_type: 'system',
  status: { status: 'in-progress', history: [] },
} as TWorkflowStep

const mockDeploy = {
  id: 'deploy-1',
  component_name: 'api-server',
  component_id: 'comp-1',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:05:00Z',
  runner_jobs: [],
} as TDeploy

export const Default = () => (
  <DeployStepDetails
    step={mockStep}
    orgId="org-123"
    deploy={mockDeploy}
    error={null}
    isLoading={false}
  />
)

export const Loading = () => (
  <DeployStepDetails
    step={mockStep}
    orgId="org-123"
    error={null}
    isLoading={true}
  />
)

export const WithError = () => (
  <DeployStepDetails
    step={mockStep}
    orgId="org-123"
    error={new Error('Failed')}
    isLoading={false}
  />
)

export const Skeleton = () => <DeployStepDetailsSkeleton />
