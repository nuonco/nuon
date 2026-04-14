import { describe, expect, test } from 'vitest'
import { completeDeployStep } from './complete-deploy-step'

describe('completeDeployStep should handle response status codes from POST onboarding/current/steps/deploy endpoint', () => {
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await completeDeployStep({ orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('status')
    expect(result).toHaveProperty('current_step')
  })
})
