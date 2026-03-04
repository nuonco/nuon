import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getAccount } from './get-account'

describe('getAccount should handle response status codes from GET account endpoint', () => {
  test('200 status', async () => {
    const account = await getAccount()
    expect(account).toHaveProperty('id')
    expect(account).toHaveProperty('email')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getAccount()).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
