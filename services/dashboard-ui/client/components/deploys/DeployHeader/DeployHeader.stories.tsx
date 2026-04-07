export default {
  title: 'Deploys/DeployHeader',
}

import { DeployHeader } from './DeployHeader'
import { DeployContext } from '@/providers/deploy-provider'

const mockDeploy = {
  id: 'deploy-1',
  org_id: 'org-1',
  install_id: 'install-1',
  component_id: 'comp-1',
  component_name: 'api-service',
  build_id: 'bld-1',
  install_deploy_type: 'deploy',
  install_workflow_id: 'wf-1',
  created_at: '2024-01-15T10:00:00Z',
  updated_at: '2024-01-15T10:05:00Z',
  status_v2: { status: 'active', status_human_description: 'Deploy running' },
  runner_jobs: [],
  oci_artifact: null,
  component_build: { vcs_connection_commit: null },
} as any

const mockInstall = {
  id: 'install-1',
  name: 'Production',
  app_id: 'app-1',
} as any

const mockComponent = {
  id: 'comp-1',
  name: 'api-service',
  type: 'helm_chart',
  app_id: 'app-1',
} as any

const mockWorkflow = {
  id: 'wf-1',
} as any

export const Default = () => (
  <DeployContext.Provider value={{ deploy: mockDeploy }}>
    <DeployHeader
      component={mockComponent}
      workflow={mockWorkflow}
      stepId="step-1"
      deploy={mockDeploy}
      install={mockInstall}
    />
  </DeployContext.Provider>
)
