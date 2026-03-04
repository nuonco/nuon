import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstalls } from './get-installs'

describe('getInstalls should handle response status codes from GET installs endpoint', () => {
  const orgId = 'test-id'

  test('200 status with pagination params', async () => {
    const result = await getInstalls({ orgId, limit: 10, offset: 0 })
    expect(Array.isArray(result)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getInstalls({ orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
