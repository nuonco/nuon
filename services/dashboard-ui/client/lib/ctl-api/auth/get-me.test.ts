import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getMe } from './get-me'

describe('getMe should handle response status codes from GET auth/me endpoint', () => {
  test('200 status', async () => {
    const me = await getMe()
    expect(me).toHaveProperty('id')
    expect(me).toHaveProperty('email')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getMe()).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
