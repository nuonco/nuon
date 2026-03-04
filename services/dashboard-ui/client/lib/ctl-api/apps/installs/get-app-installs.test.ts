import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getAppInstalls } from './get-app-installs'

describe('getAppInstalls should handle response status codes from GET /apps/:appId/installs?:params endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-app-id'

  test('200 status', async () => {
    const result = await getAppInstalls({ appId, orgId, limit: 10, offset: 0 })
    expect(Array.isArray(result)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getAppInstalls({ appId, orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
