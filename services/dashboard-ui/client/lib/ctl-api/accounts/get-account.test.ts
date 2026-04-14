import { describe, expect, test } from 'vitest'
import { getAccount } from './get-account'

describe('getAccount should handle response status codes from GET account endpoint', () => {
  test('200 status', async () => {
    const account = await getAccount()
    expect(account).toHaveProperty('id')
    expect(account).toHaveProperty('email')
  })
})
