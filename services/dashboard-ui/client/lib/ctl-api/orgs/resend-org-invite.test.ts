import { describe, expect, test } from 'vitest'
import { resendOrgInvite } from './resend-org-invite'

describe('resendOrgInvite should handle response status codes from POST orgs/current/invites/:id/resend endpoint', () => {
  const orgId = 'test-org-id'
  const inviteId = 'test-invite-id'

  test('200 status', async () => {
    const result = await resendOrgInvite({ inviteId, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('email')
    expect(result).toHaveProperty('org_id')
  })
})
