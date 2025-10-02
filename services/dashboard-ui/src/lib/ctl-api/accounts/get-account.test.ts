import '@test/mock-auth'
import { describe, expect, test } from 'vitest'
import { getAccount } from './get-account'

describe('getAccount should handle response status codes from GET account endpoint', () => {
  test('200 status', async () => {
    const { data: account } = await getAccount()
    expect(account).toHaveProperty('id')
    expect(account).toHaveProperty('email')
  })

  test.each([401, 404, 500])('%s status', async (code) => {
    const { error, status } = await getAccount()
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
