import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { resendOrgInvite } from './resend-org-invite'

describe('resendOrgInvite should handle response status codes from POST orgs/current/invites/:id/resend endpoint', () => {
  const orgId = 'test-org-id'
  const inviteId = 'test-invite-id'

  test('200 status', async () => {
    const { data: invite } = await resendOrgInvite({
      inviteId,
      orgId,
    })
    expect(invite).toHaveProperty('id')
    expect(invite).toHaveProperty('email')
    expect(invite).toHaveProperty('org_id')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await resendOrgInvite({
      inviteId,
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
