import { describe, expect, test } from 'vitest'
import { completeUserJourney } from './complete-user-journey'

describe('completeUserJourney should handle response status codes from POST account/user-journeys/:journeyName/complete endpoint', () => {
  test('200 status', async () => {
    const account = await completeUserJourney({ journeyName: 'onboarding' })
    expect(account).toHaveProperty('id')
    expect(account).toHaveProperty('email')
    expect(account).toHaveProperty('user_journeys')
  })
})
