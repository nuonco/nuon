import { describe, expect, test } from 'bun:test'
import type { TWorkflowStep } from '@/types'
import { toPendingChangeSummary } from './use-workflow-change-summaries'

const step = (over: Partial<TWorkflowStep>): TWorkflowStep => ({
  id: 'step-1',
  name: 'sync and plan api',
  execution_type: 'approval',
  ...over,
})

describe('toPendingChangeSummary', () => {
  test('synthesizes a placeholder for an unreached component plan step', () => {
    const summary = toPendingChangeSummary(
      step({ metadata: { component_name: 'api' } })
    )

    expect(summary).toEqual({
      stepId: 'step-1',
      stepName: 'sync and plan api',
      componentName: 'api',
      planType: 'pending',
      status: 'not-generated',
      counts: { create: 0, update: 0, delete: 0, replace: 0, noop: 0 },
      countsState: 'unsupported',
      hasDetail: false,
    })
  })

  test('labels an unreached sandbox plan as terraform', () => {
    const summary = toPendingChangeSummary(
      step({ name: 'provision sandbox plan' })
    )

    expect(summary?.planType).toBe('terraform_plan')
  })

  test('returns null once the step already has an approval', () => {
    const summary = toPendingChangeSummary(
      step({ approval: { id: 'appr-1', type: 'terraform_plan' } })
    )

    expect(summary).toBeNull()
  })

  test('returns null for a retried (superseded) attempt', () => {
    const summary = toPendingChangeSummary(step({ retried: true }))

    expect(summary).toBeNull()
  })

  test('returns null for steps that are not plan-approval steps', () => {
    const summary = toPendingChangeSummary(
      step({ execution_type: 'system', name: 'verify health api' })
    )

    expect(summary).toBeNull()
  })
})
