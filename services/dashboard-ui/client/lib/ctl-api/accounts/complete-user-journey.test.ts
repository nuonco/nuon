import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { completeUserJourney } from './complete-user-journey'

describe('completeUserJourney should handle response status codes from POST account/user-journeys/:journeyName/complete endpoint', () => {
  test('200 status', async () => {
    const account = await completeUserJourney({ journeyName: 'onboarding' })
    expect(account).toHaveProperty('id')
    expect(account).toHaveProperty('email')
    expect(account).toHaveProperty('user_journeys')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      completeUserJourney({ journeyName: 'test-journey' })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
