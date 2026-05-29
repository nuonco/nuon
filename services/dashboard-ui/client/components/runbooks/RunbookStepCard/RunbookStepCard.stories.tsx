import { RunbookStepCard } from './RunbookStepCard'
import type { TWorkflowStep } from '@/types'

export default {
  title: 'Runbooks/RunbookStepCard',
}

const baseStep: TWorkflowStep = {
  id: 'step-1',
  name: 'deploy api server',
  step_target_type: 'install_deploys',
  step_target_id: 'deploy-1',
  install_workflow_id: 'wf-1',
  created_by: { email: 'dev@nuon.co' },
  status: {
    status: 'succeeded',
    history: [
      { status: 'pending', created_at_ts: 1716000000 },
      { status: 'running', created_at_ts: 1716000060 },
    ],
  },
} as unknown as TWorkflowStep

const mockDeployData = {
  id: 'deploy-1',
  install_id: 'inst-1',
  component_id: 'comp-1',
  component_name: 'api-server',
  status: 'succeeded',
  created_at: '2025-05-18T10:00:00Z',
  updated_at: '2025-05-18T10:05:00Z',
}

const mockDeployOutputs = {
  cluster_endpoint: 'https://eks.us-west-2.amazonaws.com/cluster-123',
  cluster_name: 'prod-api',
  load_balancer_dns: 'api-lb-1234567.us-west-2.elb.amazonaws.com',
  namespace: 'api-server',
  service_account_arn: 'arn:aws:iam::123456789:role/api-server-sa',
}

const mockHelmDeployData = {
  id: 'deploy-2',
  install_id: 'inst-1',
  component_id: 'comp-2',
  component_name: 'monitoring',
  status: 'succeeded',
  created_at: '2025-05-18T10:01:00Z',
  updated_at: '2025-05-18T10:04:00Z',
}

const mockHelmDeployOutputs = {
  grafana: {
    url: 'https://grafana.internal:3000',
    admin_user: 'admin',
  },
  prometheus: {
    endpoint: 'http://prometheus:9090',
    retention: '30d',
  },
}

const mockActionRunData = {
  id: 'run-1',
  install_id: 'inst-1',
  action_workflow_id: 'aw-1',
  status: 'succeeded',
  config: {
    steps: [
      { id: 'sc-1', name: 'migrate', idx: 0 },
    ],
  },
  steps: [
    { id: 'rs-1', name: 'migrate', idx: 0, status: 'succeeded', execution_duration: 5000000000 },
  ],
  outputs: {
    steps: {
      migrate: { db_version: '42', rows_affected: '1024' },
    },
  },
  run_env_vars: {
    DATABASE_URL: 'postgres://db:5432/app',
    MIGRATION_DRY_RUN: 'false',
  },
  created_at: '2025-05-18T10:00:00Z',
}

const mockActionRunNoOutputs = {
  id: 'run-2',
  install_id: 'inst-1',
  action_workflow_id: 'aw-2',
  status: 'succeeded',
  config: {
    steps: [
      { id: 'sc-2', name: 'health-check', idx: 0 },
    ],
  },
  steps: [
    { id: 'rs-2', name: 'health-check', idx: 0, status: 'succeeded', execution_duration: 2000000000 },
  ],
  outputs: { steps: {} },
  created_at: '2025-05-18T11:00:00Z',
}

const mockMultiStepActionRun = {
  id: 'run-3',
  install_id: 'inst-1',
  action_workflow_id: 'aw-3',
  status: 'succeeded',
  config: {
    steps: [
      { id: 'sc-3', name: 'backup', idx: 0 },
      { id: 'sc-4', name: 'migrate', idx: 1 },
      { id: 'sc-5', name: 'seed', idx: 2 },
    ],
  },
  steps: [
    { id: 'rs-3', name: 'backup', idx: 0, status: 'succeeded', execution_duration: 3000000000 },
    { id: 'rs-4', name: 'migrate', idx: 1, status: 'succeeded', execution_duration: 8000000000 },
    { id: 'rs-5', name: 'seed', idx: 2, status: 'succeeded', execution_duration: 1000000000 },
  ],
  outputs: {
    steps: {
      backup: { snapshot_id: 'snap-abc123', size_mb: '512' },
      migrate: { version_from: '41', version_to: '42', applied_count: '3' },
      seed: { records_inserted: '150' },
    },
  },
  run_env_vars: {
    DATABASE_URL: 'postgres://db:5432/app',
    SEED_FILE: 's3://bucket/seed-data.sql',
  },
  created_at: '2025-05-18T12:00:00Z',
}

export const DeployWithOutputs = () => (
  <RunbookStepCard
    step={baseStep}
    workflowUrl="/org1/installs/inst-1/workflows/wf-1"
    targetData={mockDeployData}
    deployOutputs={mockDeployOutputs}
  />
)

export const DeployWithNestedOutputs = () => (
  <RunbookStepCard
    step={{
      ...baseStep,
      id: 'step-helm',
      name: 'deploy monitoring stack',
    } as unknown as TWorkflowStep}
    workflowUrl="/org1/installs/inst-1/workflows/wf-1"
    targetData={mockHelmDeployData}
    deployOutputs={mockHelmDeployOutputs}
  />
)

export const DeployNoOutputs = () => (
  <RunbookStepCard
    step={{
      ...baseStep,
      id: 'step-no-out',
      name: 'deploy worker',
    } as unknown as TWorkflowStep}
    workflowUrl="/org1/installs/inst-1/workflows/wf-1"
    targetData={mockDeployData}
  />
)

export const ActionRunWithOutputs = () => (
  <RunbookStepCard
    step={{
      ...baseStep,
      id: 'step-2',
      name: 'run migrations',
      step_target_type: 'install_action_workflow_runs',
      step_target_id: 'run-1',
    } as unknown as TWorkflowStep}
    workflowUrl="/org1/installs/inst-1/workflows/wf-1"
    targetData={mockActionRunData}
  />
)

export const ActionRunNoOutputs = () => (
  <RunbookStepCard
    step={{
      ...baseStep,
      id: 'step-hc',
      name: 'health check',
      step_target_type: 'install_action_workflow_runs',
      step_target_id: 'run-2',
    } as unknown as TWorkflowStep}
    workflowUrl="/org1/installs/inst-1/workflows/wf-1"
    targetData={mockActionRunNoOutputs}
  />
)

export const ActionRunMultiStep = () => (
  <RunbookStepCard
    step={{
      ...baseStep,
      id: 'step-multi',
      name: 'backup migrate and seed',
      step_target_type: 'install_action_workflow_runs',
      step_target_id: 'run-3',
    } as unknown as TWorkflowStep}
    workflowUrl="/org1/installs/inst-1/workflows/wf-1"
    targetData={mockMultiStepActionRun}
  />
)

export const Loading = () => (
  <RunbookStepCard
    step={baseStep}
    workflowUrl="/org1/installs/inst-1/workflows/wf-1"
    isLoading
  />
)

export const NoData = () => (
  <RunbookStepCard
    step={{
      ...baseStep,
      id: 'step-pending',
      name: 'pending step',
      status: { status: 'pending' },
    } as unknown as TWorkflowStep}
    workflowUrl="/org1/installs/inst-1/workflows/wf-1"
  />
)

export const FailedWithErrorDescription = () => (
  <RunbookStepCard
    step={{
      ...baseStep,
      id: 'step-failed',
      name: 'deploy database',
      status: {
        status: 'failed',
        status_human_description: 'Terraform apply failed: resource limit exceeded',
        history: [
          { status: 'pending', created_at_ts: 1716000000 },
          { status: 'running', created_at_ts: 1716000060 },
          {
            status: 'failed',
            created_at_ts: 1716000120,
            status_human_description: 'Terraform apply failed: resource limit exceeded',
          },
        ],
      },
    } as unknown as TWorkflowStep}
    workflowUrl="/org1/installs/inst-1/workflows/wf-1"
    targetData={mockDeployData}
    deployOutputs={mockDeployOutputs}
  />
)
