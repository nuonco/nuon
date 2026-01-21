import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getOrgInvites } from './get-org-invites'

describe('getOrgInvites should handle response status codes from GET orgs/current/invites endpoint', () => {
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const { data: invites } = await getOrgInvites({
      orgId,
      limit: 10,
      offset: 0,
    })

    // Validate that we got an array
    expect(Array.isArray(invites)).toBe(true)

    // Validate each invite has expected properties
    invites?.forEach((invite) => {
      expect(invite).toHaveProperty('id')
      expect(invite).toHaveProperty('email')
      expect(invite).toHaveProperty('org_id')
      expect(invite).toHaveProperty('created_at')
    })
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getOrgInvites({
      orgId,
      limit: 10,
      offset: 0,
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
