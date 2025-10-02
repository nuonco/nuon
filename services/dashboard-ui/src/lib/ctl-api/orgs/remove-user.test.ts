import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { removeUser, type TRemoveUserBody } from './remove-user'

describe('removeUser should handle response status codes from POST orgs/current/remove-user endpoint', () => {
  const orgId = 'test-org-id'
  const validBody: TRemoveUserBody = {
    user_id: 'test-user-id',
  }

  test('200 status', async () => {
    const { data: account } = await removeUser({
      body: validBody,
      orgId,
    })
    expect(account).toHaveProperty('id')
    expect(account).toHaveProperty('email')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await removeUser({
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