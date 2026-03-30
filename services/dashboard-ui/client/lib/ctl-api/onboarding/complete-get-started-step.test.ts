import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { completeGetStartedStep } from './complete-get-started-step'

describe('completeGetStartedStep should handle response status codes from POST onboarding/current/steps/get-started endpoint', () => {
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await completeGetStartedStep({ orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('status')
    expect(result).toHaveProperty('current_step')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(completeGetStartedStep({ orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
