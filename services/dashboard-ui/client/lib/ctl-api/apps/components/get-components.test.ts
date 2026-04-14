import { describe, expect, test } from 'vitest'
import { getComponents } from './get-components'

describe('getComponents should handle response status codes from GET /apps/:appId/components?:params endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-app-id'

  test('200 status', async () => {
    const result = await getComponents({ appId, orgId, limit: 10, offset: 0 })
    expect(Array.isArray(result.data)).toBe(true)
  })
})
