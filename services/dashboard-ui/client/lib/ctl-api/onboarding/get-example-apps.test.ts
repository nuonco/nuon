import { describe, expect, test } from 'vitest'
import { getExampleApps } from './get-example-apps'

describe('getExampleApps should handle response status codes from GET onboarding/example-apps endpoint', () => {
  test('200 status', async () => {
    const result = await getExampleApps()
    expect(Array.isArray(result)).toBe(true)
  })
})
