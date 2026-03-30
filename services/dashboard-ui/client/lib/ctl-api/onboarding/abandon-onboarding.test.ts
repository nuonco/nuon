import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { abandonOnboarding } from './abandon-onboarding'

describe('abandonOnboarding should handle response status codes from DELETE onboarding/current endpoint', () => {
  test('200 status', async () => {
    const result = await abandonOnboarding()
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('status')
    expect(result).toHaveProperty('current_step')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(abandonOnboarding()).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
