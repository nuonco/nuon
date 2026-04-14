import { describe, expect, test } from 'vitest'
import { getApps } from './get-apps'

describe('getApps should handle response status codes from GET apps endpoint', () => {
  const orgId = 'test-id'

  test('200 status with all optional params', async () => {
    const result = await getApps({ orgId, q: 'test-query', limit: 10, offset: 0 })
    expect(Array.isArray(result.data)).toBe(true)
  })
})
