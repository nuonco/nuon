import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { inviteUser, type TInviteUserBody } from './invite-user'

describe('inviteUser should handle response status codes from POST orgs/current/invites endpoint', () => {
  const orgId = 'test-org-id'
  const validBody: TInviteUserBody = { email: 'user@example.com' }

  test('201 status', async () => {
    const result = await inviteUser({ body: validBody, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('email')
    expect(result).toHaveProperty('org_id')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(inviteUser({ body: validBody, orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
