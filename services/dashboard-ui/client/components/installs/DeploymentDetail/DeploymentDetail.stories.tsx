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

const warn = (policy_id: string, message: string) => ({
  policy_id,
  severity: 'warn' as const,
  message,
})

const policyReportsWorkflow = {
  id: 'workflow-policy-reports',
  status: { status: 'in-progress' },
  steps: [
    {
      id: 'plan-acme-db',
      name: 'sync and plan acme-db',
      metadata: { component_name: 'acme-db' },
      status: {
        status: 'approval-awaiting',
        metadata: {
          deny_violations: [
            {
              policy_id: 'policy-protected-db',
              severity: 'deny',
              message:
                "Replacing protected resource 'module.db.aws_db_instance.primary' is blocked. Remove it from this plan before approving.",
            },
          ],
          warn_violations: [
            warn(
              'policy-db-backups',
              "Database 'module.db.aws_db_instance.primary' keeps backups for 1 day. Keep them for at least 7 days."
            ),
            warn(
              'policy-db-deletion',
              "Database 'module.db.aws_db_instance.primary' has deletion protection turned off."
            ),
          ],
        },
      },
    },
    {
      id: 'plan-acme-api',
      name: 'sync and plan acme-api',
      metadata: { component_name: 'acme-api' },
      status: {
        status: 'approval-awaiting',
        metadata: {
          warn_violations: Array.from({ length: 19 }, (_, index) => {
            const name = `acme-service-${index + 1}`
            return warn(
              `policy-non-root-${name}`,
              `Deployment '${name}' can run as root. Set securityContext.runAsNonRoot=true on the pod or container.`
            )
          }),
        },
      },
    },
  ],
} as TWorkflow

export const PolicyReports = () => {
  const metrics = useWorkflowMetrics(policyReportsWorkflow)
  return (
    <WorkflowContext.Provider
      value={{
        workflow: policyReportsWorkflow,
        ...metrics,
        stopPolling: () => {},
      }}
    >
      <div className="flex max-w-5xl flex-col gap-4 p-4">
        <DeploymentAlerts />
      </div>
    </WorkflowContext.Provider>
  )
}

export const SkippedByPolicy = () => <PolicySkipAlerts />

export const SkippedByPolicyWithWarnings = () => (
  <PolicySkipAlerts withWarnings />
)
