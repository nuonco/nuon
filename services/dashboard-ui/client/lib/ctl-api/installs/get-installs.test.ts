import { describe, expect, test } from 'vitest'
import { getInstalls } from './get-installs'

describe('getInstalls should handle response status codes from GET installs endpoint', () => {
  const orgId = 'test-id'

  test('200 status with pagination params', async () => {
    const result = await getInstalls({ orgId, limit: 10, offset: 0 })
    expect(Array.isArray(result.data)).toBe(true)
  })
})
