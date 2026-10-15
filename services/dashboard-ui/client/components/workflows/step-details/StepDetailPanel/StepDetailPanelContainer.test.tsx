import { afterEach, expect, test } from 'bun:test'
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { useState, type ReactNode } from 'react'
import { MemoryRouter } from 'react-router'
import { useSurfaces } from '@/hooks/use-surfaces'
import { SurfacesContext } from '@/providers/surfaces-provider'
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
