import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getExampleApps } from './get-example-apps'

describe('getExampleApps should handle response status codes from GET onboarding/example-apps endpoint', () => {
  test('200 status', async () => {
    const result = await getExampleApps()
    expect(Array.isArray(result)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getExampleApps()).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
