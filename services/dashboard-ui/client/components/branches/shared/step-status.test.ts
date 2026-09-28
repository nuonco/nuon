import { describe, expect, test } from 'bun:test'
import { isActiveStepStatus, stepStatusCategory } from './step-status'

describe('stepStatusCategory', () => {
  test('groups skipped and waiting steps away from a running step', () => {
    expect(stepStatusCategory('user-skipped')).toBe('pending')
    expect(stepStatusCategory('auto-skipped')).toBe('pending')
    expect(stepStatusCategory('pending')).toBe('pending')
    expect(stepStatusCategory('queued')).toBe('pending')
    expect(stepStatusCategory('in-progress')).toBe('active')
  })
})

describe('isActiveStepStatus', () => {
  test('treats a queued step as the step the run is on', () => {
    expect(isActiveStepStatus('queued')).toBe(true)
    expect(isActiveStepStatus('pending')).toBe(false)
    expect(isActiveStepStatus('in-progress')).toBe(true)
    expect(isActiveStepStatus('user-skipped')).toBe(false)
  })
})
