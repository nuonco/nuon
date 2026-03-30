import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { completeYourStackStep } from './complete-your-stack-step'

describe('completeYourStackStep should handle response status codes from POST onboarding/current/steps/your-stack endpoint', () => {
  const orgId = 'test-org-id'
  const body = { app_type: 'example' as const, example_app_slug: 'test-app' }

  test('200 status', async () => {
    const result = await completeYourStackStep({ body, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('status')
    expect(result).toHaveProperty('current_step')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(completeYourStackStep({ body, orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
