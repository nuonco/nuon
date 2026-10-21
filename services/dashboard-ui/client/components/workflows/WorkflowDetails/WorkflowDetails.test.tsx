import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { OrgContext } from '@/providers/org-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'
import { WorkflowContext } from '@/providers/workflow-provider'
import type { TWorkflow, TWorkflowStep } from '@/types'
import { WorkflowAlertBanners } from './WorkflowDetails'

afterEach(cleanup)

const failedStep = {
  id: 'step-1',
  name: 'generate install stack',
  install_workflow_id: 'wf-1',
  execution_type: 'system',
  step_target_type: 'install_stack_versions',
  status: { status: 'error', status_human_description: 'Failed' },
} as TWorkflowStep

const workflow = {
  id: 'wf-1',
  type: 'app_branch_config_update',
  owner_id: 'ins-1',
  metadata: { install_id: 'ins-1' },
  status: { status: 'error' },
} as TWorkflow

const workflowContext = {
  workflow,
  stopPolling: () => {},
  workflowSteps: [],
  hasApprovals: false,
  failedSteps: [failedStep],
  pendingApprovals: [],
  discardedSteps: [],
  completedSteps: [],
  stepsWithPolicyViolations: [],
  totalSteps: 0,
  pendingApprovalsCount: 0,
  discardedStepsCount: 0,
  completedStepsCount: 0,
  failedStepsCount: 0,
  policyViolationsCount: 0,
}

test('a failed-step banner opens its step panel from a page with no install provider above it', async () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: Infinity } },
  })
  client.setQueryData(['install', 'org-1', 'ins-1'], { id: 'ins-1' })

  render(
    <QueryClientProvider client={client}>
      <OrgContext.Provider
        value={{ org: { id: 'org-1' } as never, refresh: () => {} }}
      >
        <ToastProvider>
          <MemoryRouter initialEntries={['/org-1/apps/app-1/branches/br-1']}>
            <SurfacesProvider>
              <WorkflowContext.Provider value={workflowContext}>
                <WorkflowAlertBanners
                  workflow={workflow}
                  failedSteps={[failedStep]}
                />
              </WorkflowContext.Provider>
            </SurfacesProvider>
          </MemoryRouter>
        </ToastProvider>
      </OrgContext.Provider>
    </QueryClientProvider>
  )

  fireEvent.click(screen.getByText('View details'))

  const panels = await screen.findAllByRole('complementary')
  expect(panels).toHaveLength(1)
})
