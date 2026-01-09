import '@test/mock-auth'
import { describe, expect, test } from 'vitest'
import { getMe } from './get-me'

describe('getMe should handle response status codes from GET auth/me endpoint', () => {
  test('200 status', async () => {
    const { data: me } = await getMe()
    expect(me).toHaveProperty('id')
    expect(me).toHaveProperty('email')
  })

  test.each([401, 500])('%s status', async (code) => {
    const { error, status } = await getMe()
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})