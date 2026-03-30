import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getCurrentOnboarding } from './get-current-onboarding'

describe('getCurrentOnboarding should handle response status codes from GET onboarding/current endpoint', () => {
  test('200 status', async () => {
    const result = await getCurrentOnboarding()
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('status')
    expect(result).toHaveProperty('current_step')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getCurrentOnboarding()).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
