export default {
  title: 'Workflows/ActiveWorkflows',
}

import type { ReactNode } from 'react'
import { WorkflowApprovalsContext } from '@/providers/workflow-approvals-provider'
import { ActiveWorkflows } from './ActiveWorkflows'

function mockWorkflow(
  id: string,
  type: string,
  name: string,
  installName: string,
  minutesAgo = 0,
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
  }
}

const approvalsFor = (...workflowIds: string[]) =>
  workflowIds.map((id) => ({
    id: `approval-${id}`,
    workflow_step: { id: `step-${id}`, install_workflow_id: id },
  })) as any

const ApprovalsProvider = ({
  approvals = [],
  children,
}: {
  approvals?: any[]
  children: ReactNode
}) => (
  <WorkflowApprovalsContext.Provider
    value={{ approvals, isLoading: false, refresh: () => {} }}
  >
    {children}
  </WorkflowApprovalsContext.Provider>
)

const one = [
  mockWorkflow('wf-1', 'deploy_components', 'Deploy components', 'My Install'),
] as any

const two = [
  mockWorkflow('wf-1', 'deploy_components', 'Deploy components', 'My Install'),
  mockWorkflow('wf-2', 'provision_sandbox', 'Provision sandbox', 'Staging', 5),
] as any

const four = [
  mockWorkflow('wf-1', 'deploy_components', 'Deploy components', 'My Install'),
  mockWorkflow('wf-2', 'provision_sandbox', 'Provision sandbox', 'Staging', 5),
  mockWorkflow('wf-3', 'action_workflow_run', 'Run migrations', 'Production', 10),
  mockWorkflow('wf-4', 'reprovision_sandbox', 'Reprovision sandbox', 'Dev', 2),
] as any

const six = [
  mockWorkflow('wf-1', 'deploy_components', 'Deploy components', 'My Install'),
  mockWorkflow('wf-2', 'provision_sandbox', 'Provision sandbox', 'Staging', 5),
  mockWorkflow('wf-3', 'action_workflow_run', 'Run migrations', 'Production', 10),
  mockWorkflow('wf-4', 'reprovision_sandbox', 'Reprovision sandbox', 'Dev', 2),
  mockWorkflow('wf-5', 'deploy_components', 'Deploy components', 'Customer A', 8),
  mockWorkflow('wf-6', 'drift_run', 'Drift run', 'Customer B', 15),
] as any

const many = [
  mockWorkflow('wf-1', 'deploy_components', 'Deploy components', 'My Install'),
  mockWorkflow('wf-2', 'provision_sandbox', 'Provision sandbox', 'Staging', 5),
  mockWorkflow('wf-3', 'action_workflow_run', 'Run migrations', 'Production', 10),
  mockWorkflow('wf-4', 'reprovision_sandbox', 'Reprovision sandbox', 'Dev', 2),
  mockWorkflow('wf-5', 'deploy_components', 'Deploy components', 'Customer A', 8),
  mockWorkflow('wf-6', 'drift_run', 'Drift run', 'Customer B', 15),
  mockWorkflow('wf-7', 'deploy_components', 'Deploy all', 'Customer C', 1),
  mockWorkflow('wf-8', 'provision_sandbox', 'Provision sandbox', 'Customer D', 3),
  mockWorkflow('wf-9', 'action_workflow_run', 'Sync secrets', 'Customer E', 12),
  mockWorkflow('wf-10', 'deploy_components', 'Deploy components', 'Customer F', 20),
] as any

export const Single = () => (
  <ApprovalsProvider>
    <div className="p-4">
      <ActiveWorkflows workflows={one} />
    </div>
  </ApprovalsProvider>
)

export const Two = () => (
  <ApprovalsProvider>
    <div className="p-4">
      <ActiveWorkflows workflows={two} />
    </div>
  </ApprovalsProvider>
)

export const Four = () => (
  <ApprovalsProvider approvals={approvalsFor('wf-2', 'wf-4')}>
    <div className="p-4">
      <ActiveWorkflows workflows={four} />
    </div>
  </ApprovalsProvider>
)

export const Six = () => (
  <ApprovalsProvider approvals={approvalsFor('wf-2', 'wf-5')}>
    <div className="p-4">
      <ActiveWorkflows workflows={six} />
    </div>
  </ApprovalsProvider>
)

export const Ten = () => (
  <ApprovalsProvider approvals={approvalsFor('wf-2', 'wf-4', 'wf-6', 'wf-8')}>
    <div className="p-4">
      <ActiveWorkflows workflows={many} />
    </div>
  </ApprovalsProvider>
)

export const Empty = () => (
  <ApprovalsProvider>
    <div className="p-4">
      <ActiveWorkflows workflows={[]} />
    </div>
  </ApprovalsProvider>
)
