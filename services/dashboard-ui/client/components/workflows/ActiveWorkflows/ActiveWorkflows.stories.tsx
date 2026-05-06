export default {
  title: 'Workflows/ActiveWorkflows',
}

import { ActiveWorkflows } from './ActiveWorkflows'

const pendingStep = {
  execution_type: 'approval',
  status: { status: 'approval-awaiting' },
  approval: { response: null },
}

function mockWorkflow(
  id: string,
  type: string,
  name: string,
  installName: string,
  minutesAgo = 0,
  opts?: { pendingApproval?: boolean },
) {
  return {
    id,
    type,
    name,
    owner_id: `install-${id}`,
    status: { status: 'in-progress' },
    metadata: { owner_name: installName },
    created_at: new Date(Date.now() - minutesAgo * 60000).toISOString(),
    updated_at: new Date().toISOString(),
    steps: opts?.pendingApproval ? [pendingStep] : [],
  }
}

const one = [
  mockWorkflow('wf-1', 'deploy_components', 'Deploy components', 'My Install'),
] as any

const two = [
  mockWorkflow('wf-1', 'deploy_components', 'Deploy components', 'My Install'),
  mockWorkflow('wf-2', 'provision_sandbox', 'Provision sandbox', 'Staging', 5),
] as any

const four = [
  mockWorkflow('wf-1', 'deploy_components', 'Deploy components', 'My Install'),
  mockWorkflow('wf-2', 'provision_sandbox', 'Provision sandbox', 'Staging', 5, { pendingApproval: true }),
  mockWorkflow('wf-3', 'action_workflow_run', 'Run migrations', 'Production', 10),
  mockWorkflow('wf-4', 'reprovision_sandbox', 'Reprovision sandbox', 'Dev', 2, { pendingApproval: true }),
] as any

const six = [
  mockWorkflow('wf-1', 'deploy_components', 'Deploy components', 'My Install'),
  mockWorkflow('wf-2', 'provision_sandbox', 'Provision sandbox', 'Staging', 5, { pendingApproval: true }),
  mockWorkflow('wf-3', 'action_workflow_run', 'Run migrations', 'Production', 10),
  mockWorkflow('wf-4', 'reprovision_sandbox', 'Reprovision sandbox', 'Dev', 2),
  mockWorkflow('wf-5', 'deploy_components', 'Deploy components', 'Customer A', 8, { pendingApproval: true }),
  mockWorkflow('wf-6', 'drift_run', 'Drift run', 'Customer B', 15),
] as any

const many = [
  mockWorkflow('wf-1', 'deploy_components', 'Deploy components', 'My Install'),
  mockWorkflow('wf-2', 'provision_sandbox', 'Provision sandbox', 'Staging', 5, { pendingApproval: true }),
  mockWorkflow('wf-3', 'action_workflow_run', 'Run migrations', 'Production', 10),
  mockWorkflow('wf-4', 'reprovision_sandbox', 'Reprovision sandbox', 'Dev', 2, { pendingApproval: true }),
  mockWorkflow('wf-5', 'deploy_components', 'Deploy components', 'Customer A', 8),
  mockWorkflow('wf-6', 'drift_run', 'Drift run', 'Customer B', 15, { pendingApproval: true }),
  mockWorkflow('wf-7', 'deploy_components', 'Deploy all', 'Customer C', 1),
  mockWorkflow('wf-8', 'provision_sandbox', 'Provision sandbox', 'Customer D', 3, { pendingApproval: true }),
  mockWorkflow('wf-9', 'action_workflow_run', 'Sync secrets', 'Customer E', 12),
  mockWorkflow('wf-10', 'deploy_components', 'Deploy components', 'Customer F', 20),
] as any

export const Single = () => (
  <div className="p-4">
    <ActiveWorkflows workflows={one} />
  </div>
)

export const Two = () => (
  <div className="p-4">
    <ActiveWorkflows workflows={two} />
  </div>
)

export const Four = () => (
  <div className="p-4">
    <ActiveWorkflows workflows={four} />
  </div>
)

export const Six = () => (
  <div className="p-4">
    <ActiveWorkflows workflows={six} />
  </div>
)

export const Ten = () => (
  <div className="p-4">
    <ActiveWorkflows workflows={many} />
  </div>
)

export const Empty = () => (
  <div className="p-4">
    <ActiveWorkflows workflows={[]} />
  </div>
)
