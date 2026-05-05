export default {
  title: 'Workflows/WorkflowDetails',
}

import type { ReactNode } from 'react'
import { WorkflowDetails } from './WorkflowDetails'
import { WorkflowContext } from '@/providers/workflow-provider'
import { OrgContext } from '@/providers/org-provider'
import { InstallContext } from '@/providers/install-provider'
import type { TWorkflow, TWorkflowStep } from '@/types'

const mockOrg = { id: 'org-1', name: 'Acme Corp' } as any
const mockInstall = { id: 'inst-1', name: 'production', app_id: 'app-1' } as any

const baseWorkflow = {
  id: 'wf-123',
  name: 'deploy',
  type: 'deploy_components',
  created_at: '2026-05-01T10:00:00Z',
  updated_at: '2026-05-01T10:05:00Z',
  status: {
    status: 'success',
    status_human_description: 'Succeeded',
    metadata: {},
  },
  steps: [],
} as TWorkflow

const makeStep = (
  id: string,
  name: string,
  status: string,
  statusDescription?: string
) =>
  ({
    id,
    name,
    execution_type: 'system',
    status: {
      status,
      status_human_description: statusDescription ?? status,
      history: [],
    },
  }) as TWorkflowStep

const Wrapper = ({
  children,
  workflow,
}: {
  children: ReactNode
  workflow: TWorkflow
}) => (
  <OrgContext.Provider value={{ org: mockOrg, refresh: () => {} }}>
    <InstallContext.Provider value={{ install: mockInstall, refresh: () => {} }}>
      <WorkflowContext.Provider
        value={{
          workflow,
          stopPolling: () => {},
          workflowSteps: workflow.steps ?? [],
          hasApprovals: false,
          failedSteps: [],
          pendingApprovals: [],
          discardedSteps: [],
          completedSteps: [],
          stepsWithPolicyViolations: [],
          totalSteps: workflow.steps?.length ?? 0,
          pendingApprovalsCount: 0,
          discardedStepsCount: 0,
          completedStepsCount: 0,
          failedStepsCount: 0,
          policyViolationsCount: 0,
        }}
      >
        {children}
      </WorkflowContext.Provider>
    </InstallContext.Provider>
  </OrgContext.Provider>
)

export const Default = () => (
  <Wrapper workflow={baseWorkflow}>
    <WorkflowDetails workflow={baseWorkflow} failedSteps={[]} />
  </Wrapper>
)

export const RetriesExhausted = () => {
  const workflow = {
    ...baseWorkflow,
    status: {
      ...baseWorkflow.status,
      status: 'error',
      status_human_description: 'Failed',
      metadata: { retries_exhausted: true, max_retries: 3 },
    },
  } as TWorkflow

  return (
    <Wrapper workflow={workflow}>
      <WorkflowDetails workflow={workflow} failedSteps={[]} />
    </Wrapper>
  )
}

export const WorkflowStopped = () => {
  const workflow = {
    ...baseWorkflow,
    status: {
      ...baseWorkflow.status,
      status: 'error',
      status_human_description: 'Stopped',
      metadata: {
        stopped: true,
        error_message: 'Workflow was cancelled by user before completion.',
      },
    },
  } as TWorkflow

  return (
    <Wrapper workflow={workflow}>
      <WorkflowDetails workflow={workflow} failedSteps={[]} />
    </Wrapper>
  )
}

export const SingleFailedStep = () => {
  const workflow = {
    ...baseWorkflow,
    status: {
      ...baseWorkflow.status,
      status: 'error',
      status_human_description: 'Failed',
      metadata: {},
    },
  } as TWorkflow

  return (
    <Wrapper workflow={workflow}>
      <WorkflowDetails
        workflow={workflow}
        failedSteps={[
          makeStep(
            'step-1',
            'deploy api-server',
            'error',
            'Terraform apply failed: resource limit exceeded'
          ),
        ]}
      />
    </Wrapper>
  )
}

export const MultipleFailedSteps = () => {
  const workflow = {
    ...baseWorkflow,
    status: {
      ...baseWorkflow.status,
      status: 'error',
      status_human_description: 'Failed',
      metadata: {},
    },
  } as TWorkflow

  return (
    <Wrapper workflow={workflow}>
      <WorkflowDetails
        workflow={workflow}
        failedSteps={[
          makeStep('step-1', 'deploy database', 'error', 'Connection timeout after 300s'),
          makeStep('step-2', 'deploy cache', 'error', 'Helm release failed: image pull backoff'),
          makeStep(
            'step-3',
            'deploy api-server',
            'error',
            'Terraform apply failed: resource limit exceeded'
          ),
        ]}
      />
    </Wrapper>
  )
}

export const StoppedWithFailedSteps = () => {
  const workflow = {
    ...baseWorkflow,
    status: {
      ...baseWorkflow.status,
      status: 'error',
      status_human_description: 'Stopped',
      metadata: {
        stopped: true,
        error_message: 'Workflow stopped due to step failure.',
      },
    },
  } as TWorkflow

  return (
    <Wrapper workflow={workflow}>
      <WorkflowDetails
        workflow={workflow}
        failedSteps={[
          makeStep(
            'step-1',
            'deploy api-server',
            'error',
            'Terraform plan failed: invalid provider configuration'
          ),
        ]}
      />
    </Wrapper>
  )
}
