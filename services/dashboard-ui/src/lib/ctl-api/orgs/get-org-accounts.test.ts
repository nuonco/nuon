import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getOrgAccounts } from './get-org-accounts'

describe('getOrgAccounts should handle response status codes from GET endpoint', () => {
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const { data } = await getOrgAccounts({
      orgId,
      limit: 10,
      offset: 0,
    })
    expect(data).toHaveProperty('id')
    expect(data).toHaveProperty('account_type')
    expect(data).toHaveProperty('email')
    expect(data).toHaveProperty('created_at')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getOrgAccounts({
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
