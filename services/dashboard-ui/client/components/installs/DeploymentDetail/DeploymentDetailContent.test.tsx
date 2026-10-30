import { expect, test } from 'bun:test'
import { render, screen } from '@testing-library/react'
import { useWorkflowMetrics } from '@/hooks/use-workflow-metrics'
import { WorkflowContext } from '@/providers/workflow-provider'
import type { TWorkflow } from '@/types'
import { DeploymentAlerts } from './DeploymentDetailContent'

const Alerts = ({ workflow }: { workflow: TWorkflow }) => {
  const metrics = useWorkflowMetrics(workflow)
  return (
    <WorkflowContext.Provider
      value={{ workflow, ...metrics, stopPolling: () => {} }}
    >
      <DeploymentAlerts />
    </WorkflowContext.Provider>
  )
}

test.each([false, true])(
  'shows the policy skip decision with warnings present: %s',
  (withWarnings) => {
    const workflow = {
      id: 'workflow-skip',
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
                  message:
                    'Sandbox apply deferred until the maintenance window.',
                },
              ],
              warn_violations: withWarnings
                ? [
                    {
                      policy_id: 'policy-warn',
                      message: 'Review instance size.',
                    },
                  ]
                : [],
              passed_policy_ids: ['policy-tags'],
            },
          },
        },
      ],
    } as TWorkflow

    render(<Alerts workflow={workflow} />)

    expect(screen.getAllByText('Policy report')).toHaveLength(1)
    expect(screen.getByText('Apply skipped by policy')).toBeDefined()
    expect(
      screen.getByText('Sandbox apply deferred until the maintenance window.')
    ).toBeDefined()
    expect(screen.queryByText('All passed')).toBeNull()
    expect(screen.queryByText('Review instance size.') !== null).toBe(
      withWarnings
    )
  }
)
