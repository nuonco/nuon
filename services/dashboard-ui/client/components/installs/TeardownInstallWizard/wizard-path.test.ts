import { describe, expect, test } from 'bun:test'
import { teardownWizardStatuses } from './wizard-path'

describe('teardown install wizard', () => {
  test('marks completed, current, and upcoming steps from the frontier', () => {
    expect(teardownWizardStatuses(1)).toEqual(['done', 'current', 'upcoming'])
    expect(teardownWizardStatuses(2)).toEqual(['done', 'done', 'current'])
  })
})
