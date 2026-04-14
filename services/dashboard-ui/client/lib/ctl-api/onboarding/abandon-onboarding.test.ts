import { describe, expect, test } from 'vitest'
import { abandonOnboarding } from './abandon-onboarding'

describe('abandonOnboarding should handle response status codes from DELETE onboarding/current endpoint', () => {
  test('200 status', async () => {
    const result = await abandonOnboarding()
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('status')
    expect(result).toHaveProperty('current_step')
  })
})
