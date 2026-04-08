export default {
  title: 'Deploys/ManagementDropdown',
}

import { ManagementDropdown } from './ManagementDropdown'
import type { TComponent, TDeploy, TWorkflow } from '@/types'

const mockComponent = {
  id: 'comp-123',
  name: 'API server',
} as TComponent

const mockDeploy = {
  id: 'dep-123',
  status_v2: { status: 'active' },
  runner_jobs: [
    { id: 'rj-1' },
    { id: 'rj-2' },
  ],
} as TDeploy

const mockWorkflow = {
  id: 'wf-123',
  name: 'deploy',
  type: 'deploy_components',
  finished: false,
} as TWorkflow

export const Default = () => (
  <ManagementDropdown
    component={mockComponent}
    currentBuildId="bld-123"
    workflow={mockWorkflow}
    deploy={mockDeploy}
  />
)

export const FinishedWorkflow = () => (
  <ManagementDropdown
    component={mockComponent}
    currentBuildId="bld-123"
    workflow={{ ...mockWorkflow, finished: true } as TWorkflow}
    deploy={mockDeploy}
  />
)
