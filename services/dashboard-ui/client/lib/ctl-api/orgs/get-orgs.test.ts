import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getOrgs } from './get-orgs'

describe('getOrgs should handle response status codes from GET orgs endpoint', () => {
  test('200 status', async () => {
    const result = await getOrgs()
    expect(Array.isArray(result)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getOrgs()).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
