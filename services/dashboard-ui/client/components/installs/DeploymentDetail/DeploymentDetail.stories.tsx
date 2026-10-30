export default {
  title: 'Features / Installs / Deployment details',
}

import { DeploymentsPreview } from './DeploymentDetail.preview'
import { useWorkflowMetrics } from '@/hooks/use-workflow-metrics'
import { WorkflowContext } from '@/providers/workflow-provider'
import type { TWorkflow } from '@/types'
import { DeploymentAlerts } from './DeploymentDetailContent'
import { ResourceScopeSummary } from './DeploymentProgress'
import { deploymentOutcomes } from './deployment-progress'

export const RunningComponents = () => (
  <DeploymentsPreview scenario="running-components" />
)

export const ComponentsNotStarted = () => (
  <DeploymentsPreview scenario="components-not-started" />
)

export const PendingApproval = () => (
  <DeploymentsPreview scenario="pending-approval" />
)

export const PartialFailure = () => (
  <DeploymentsPreview scenario="partial-failure" />
)

export const StackRecovery = () => (
  <DeploymentsPreview scenario="stack-recovery" />
)

export const Completed = () => <DeploymentsPreview scenario="completed" />

const PolicySkipAlerts = ({ withWarnings = false }) => {
  const workflow = {
    id: 'workflow-policy-skip',
    status: { status: 'success' },
    steps: [
      {
        id: 'sandbox-plan',
        name: 'reprovision sandbox plan',
        status: {
          status: 'auto-skipped',
          metadata: {
            check: 'policy-skip-apply',
            skip_findings: [
              {
                policy_id: 'policy-skip',
                message: 'Sandbox apply deferred until the maintenance window.',
              },
            ],
            warn_violations: withWarnings
              ? [{ policy_id: 'policy-warn', message: 'Review instance size.' }]
              : [],
          },
        },
      },
    ],
  } as TWorkflow
  const metrics = useWorkflowMetrics(workflow)
  return (
    <WorkflowContext.Provider
      value={{ workflow, ...metrics, stopPolling: () => {} }}
    >
      <ResourceScopeSummary
        run={{
          status: 'success',
          outcomes: deploymentOutcomes(undefined, workflow),
        }}
      />
      <DeploymentAlerts />
    </WorkflowContext.Provider>
  )
}

export const SkippedByPolicy = () => <PolicySkipAlerts />

export const SkippedByPolicyWithWarnings = () => (
  <PolicySkipAlerts withWarnings />
)
