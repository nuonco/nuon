import { afterEach, expect, test } from 'bun:test'
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useState, type ReactNode } from 'react'
import { MemoryRouter } from 'react-router'
import { Button } from '@/components/common/Button'
import { Panel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import { InstallContext } from '@/providers/install-provider'
import { OrgContext } from '@/providers/org-provider'
import { SurfacesContext, SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'
import { WorkflowContext } from '@/providers/workflow-provider'
import type { TWorkflow, TWorkflowStep } from '@/types'
import { StepDetailPanelButton } from './StepDetailPanelContainer'

afterEach(cleanup)

const PanelCount = () => {
  const { panels } = useSurfaces()
  return <div data-testid="panel-count">{panels.length}</div>
}

const workflow = {
  id: 'wf-1',
  type: 'deploy_components',
  approval_option: 'prompt',
  status: { status: 'in-progress' },
} as TWorkflow

const runningStep = {
  id: 'step-1',
  install_workflow_id: 'wf-1',
  name: 'Plan api',
  execution_type: 'approval',
  finished: false,
  status: { status: 'in-progress' },
} as TWorkflowStep

const awaitingFinishedStep = {
  ...runningStep,
  finished: true,
  status: { status: 'approval-awaiting' },
  approval: { id: 'approval-1', type: 'helm_approval' },
} as TWorkflowStep

const workflowContext = {
  workflow,
  stopPolling: () => {},
  workflowSteps: [],
  hasApprovals: false,
  failedSteps: [],
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

const MockSurfaces = ({ children }: { children: ReactNode }) => {
  const [panels, setPanels] = useState<
    Array<{
      id: string
      content: never
      isVisible: boolean
      pathname: string
    }>
  >([])
  const addPanel = () => {
    const id = `panel-${panels.length + 1}`
    setPanels((current) => [
      ...current,
      { id, content: null as never, isVisible: true, pathname: '/' },
    ])
    return id
  }

  return (
    <SurfacesContext.Provider
      value={
        {
          panels,
          modals: [],
          addPanel,
          addPanelIfEmpty: () => (panels.length === 0 ? addPanel() : undefined),
          updatePanel: () => {},
          clearPanels: () => setPanels([]),
          removePanel: (id: string) =>
            setPanels((current) => current.filter((panel) => panel.id !== id)),
          addModal: () => '',
          removeModal: () => {},
        } as never
      }
    >
      {children}
    </SurfacesContext.Provider>
  )
}

const Harness = ({ step }: { step: TWorkflowStep }) => (
  <MemoryRouter>
    <WorkflowContext.Provider value={workflowContext}>
      <MockSurfaces>
        <PanelCount />
        <StepDetailPanelButton step={step} approvalPrompt planOnly={false} />
      </MockSurfaces>
    </WorkflowContext.Provider>
  </MemoryRouter>
)

test('an approval transition does not close a manually opened step panel', async () => {
  const view = render(<Harness step={runningStep} />)

  fireEvent.click(screen.getByText('View details'))
  expect(screen.getByTestId('panel-count')).toHaveTextContent('1')

  view.rerender(<Harness step={awaitingFinishedStep} />)

  await waitFor(() =>
    expect(screen.getByTestId('panel-count')).toHaveTextContent('1')
  )
})

const OpenDeploymentPanel = () => {
  const { addPanel } = useSurfaces()
  return (
    <Button
      onClick={() =>
        addPanel(
          <Panel heading="Deployment">
            <InstallContext.Provider
              value={{
                install: { id: 'ins-1' } as never,
                labelColors: {},
                refresh: () => {},
              }}
            >
              <WorkflowContext.Provider value={workflowContext}>
                <StepDetailPanelButton step={runningStep} />
              </WorkflowContext.Provider>
            </InstallContext.Provider>
          </Panel>
        )
      }
    >
      Open deployment
    </Button>
  )
}

test('view details stacks the step panel over a workflow panel outside the install', async () => {
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
              <OpenDeploymentPanel />
            </SurfacesProvider>
          </MemoryRouter>
        </ToastProvider>
      </OrgContext.Provider>
    </QueryClientProvider>
  )

  fireEvent.click(screen.getByText('Open deployment'))
  fireEvent.click(await screen.findByText('View details'))

  expect(await screen.findByText('View workflows')).toBeTruthy()
  expect(screen.getAllByRole('complementary')).toHaveLength(2)
})
