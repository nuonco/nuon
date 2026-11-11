import { describe, expect, test } from 'bun:test'
import type { TInstallWorkflowStep } from '@/types'
import {
  filterWorkflowPanelSteps,
  parseWorkflowRunPanelKey,
  workflowRunPanelKey,
} from './WorkflowRunPanelContainer'

test('workflow run panel keys round trip', () => {
  expect(parseWorkflowRunPanelKey(workflowRunPanelKey('inw123'))).toBe('inw123')
  expect(parseWorkflowRunPanelKey('runbook-history-rb1')).toBeNull()
})

describe('filterWorkflowPanelSteps', () => {
  test('hides internal steps from the horizontal pipeline', () => {
    const steps = [
      { id: 'ignored', name: 'check ignored changes' },
      { id: 'preview', name: 'setup preview' },
      { id: 'impact', name: 'preview install impact' },
      { id: 'hidden', name: 'internal', execution_type: 'hidden' },
      { id: 'component', name: 'component', owner_type: 'components' },
      { id: 'visible', name: 'building components and sandbox' },
    ] as TInstallWorkflowStep[]

    expect(filterWorkflowPanelSteps(steps).map((step) => step.id)).toEqual([
      'visible',
    ])
  })
})
