import { describe, expect, test } from 'vitest'
import { getAppInstalls } from './get-app-installs'

describe('getAppInstalls should handle response status codes from GET /apps/:appId/installs?:params endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-app-id'

  test('200 status', async () => {
    const result = await getAppInstalls({ appId, orgId, limit: 10, offset: 0 })
    expect(Array.isArray(result.data)).toBe(true)
  })
})
