import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { inviteUser, type TInviteUserBody } from './invite-user'

describe('inviteUser should handle response status codes from POST orgs/current/invites endpoint', () => {
  const orgId = 'test-org-id'
  const validBody: TInviteUserBody = {
    email: 'user@example.com',
  }

  test('201 status', async () => {
    const { data: invite } = await inviteUser({
      body: validBody,
      orgId,
    })
    expect(invite).toHaveProperty('id')
    expect(invite).toHaveProperty('email')
    expect(invite).toHaveProperty('org_id')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await inviteUser({
      body: validBody,
      orgId,
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})