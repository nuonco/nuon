export default {
  title: 'Actions/InstallActionRunHeader',
}

import { Button } from '@/components/common/Button'
import { InstallActionRunHeader } from './InstallActionRunHeader'

const mockRun = {
  id: 'run-1',
  updated_at: new Date().toISOString(),
  created_at: new Date().toISOString(),
  triggered_by_type: 'manual',
  install_action_workflow_id: 'iaw-1',
  created_by: { email: 'user@example.com' },
  created_by_id: 'user-1',
  execution_time: 90000000000,
  config: { timeout: 300000000000 },
  status_v2: { status: 'succeeded', status_human_description: 'Completed successfully' },
  run_env_vars: { COMPONENT_NAME: 'web-app', COMPONENT_ID: 'comp-1' },
  runner_job: { id: 'job-1', install_role_usage: { role_name: 'arn:aws:iam::role/deploy' } },
} as any

const mockWorkflow = {
  id: 'wf-1',
  name: 'deploy',
  type: 'action',
} as any

export const Default = () => (
  <InstallActionRunHeader
    actionId="action-1"
    actionName="deploy-step"
    workflow={mockWorkflow}
    installActionRun={mockRun}
    basePath="/org-1/installs/install-1"
    isAdmin={false}
    step={{ id: 'step-1' } as any}
    cancelWorkflowButton={<Button variant="danger">Cancel workflow</Button>}
    runnerJobPlanButton={<Button>View plan</Button>}
  />
)

export const Admin = () => (
  <InstallActionRunHeader
    actionId="action-1"
    actionName="deploy-step"
    workflow={mockWorkflow}
    installActionRun={mockRun}
    basePath="/org-1/installs/install-1"
    isAdmin={true}
    step={{ id: 'step-1' } as any}
    cancelWorkflowButton={<Button variant="danger">Cancel workflow</Button>}
    runnerJobPlanButton={<Button>View plan</Button>}
  />
)

export const NoWorkflow = () => (
  <InstallActionRunHeader
    actionId="action-1"
    actionName="deploy-step"
    workflow={null as any}
    installActionRun={mockRun}
    basePath="/org-1/installs/install-1"
    isAdmin={false}
    cancelWorkflowButton={null}
  />
)
