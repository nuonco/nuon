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
})
