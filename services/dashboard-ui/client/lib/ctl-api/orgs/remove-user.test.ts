import { describe, expect, test } from 'vitest'
import { removeUser, type TRemoveUserBody } from './remove-user'

describe('removeUser should handle response status codes from POST orgs/current/remove-user endpoint', () => {
  const orgId = 'test-org-id'
  const validBody: TRemoveUserBody = { user_id: 'test-user-id' }

  test('200 status', async () => {
    const result = await removeUser({ body: validBody, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('email')
  })
})
