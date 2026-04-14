import { describe, expect, test } from 'vitest'
import { completeOrganizationStep } from './complete-organization-step'

describe('completeOrganizationStep should handle response status codes from POST onboarding/current/steps/organization endpoint', () => {
  const body = { name: 'Test Org' }

  test('200 status', async () => {
    const result = await completeOrganizationStep({ body })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('status')
    expect(result).toHaveProperty('current_step')
  })
})
