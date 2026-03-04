import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getOrgAccounts } from './get-org-accounts'

describe('getOrgAccounts should handle response status codes from GET endpoint', () => {
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getOrgAccounts({ orgId, limit: 10, offset: 0 })
    expect(Array.isArray(result)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      getOrgAccounts({ orgId, limit: 10, offset: 0 })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
