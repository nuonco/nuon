import { describe, expect, test } from 'bun:test'
import type { TInstallWorkflowStep } from '@/types'
import { filterWorkflowPanelSteps } from './WorkflowRunPanelContainer'

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
