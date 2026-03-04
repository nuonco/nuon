import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getApp } from './get-app'

describe('getApp should handle response status codes from GET apps/:id endpoint', () => {
  const appId = 'test-id'
  const orgId = 'test-id'

  test('200 status', async () => {
    const result = await getApp({ appId, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('name')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getApp({ appId, orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
