import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getActions } from './get-actions'

describe('getActions should handle response status codes from GET /apps/:appId/action-workflows?:params endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-app-id'

  test('200 status', async () => {
    const result = await getActions({ appId, orgId, limit: 10, offset: 0 })
    expect(Array.isArray(result)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getActions({ appId, orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
