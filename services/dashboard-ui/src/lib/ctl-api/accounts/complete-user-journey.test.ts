import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { completeUserJourney } from './complete-user-journey'

describe('completeUserJourney should handle response status codes from POST account/user-journeys/:journeyName/complete endpoint', () => {
  test('200 status', async () => {
    const { data: account } = await completeUserJourney({
      journeyName: 'onboarding',
    })
    expect(account).toHaveProperty('id')
    expect(account).toHaveProperty('email')
    expect(account).toHaveProperty('user_journeys')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await completeUserJourney({
      journeyName: 'test-journey',
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
