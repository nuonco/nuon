import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getOrgInvites } from './get-org-invites'

describe('getOrgInvites should handle response status codes from GET orgs/current/invites endpoint', () => {
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getOrgInvites({ orgId, limit: 10, offset: 0 })
    expect(Array.isArray(result)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      getOrgInvites({ orgId, limit: 10, offset: 0 })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
