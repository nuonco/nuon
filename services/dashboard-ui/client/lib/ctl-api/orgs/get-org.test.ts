import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getOrg } from './get-org'

describe('getOrg should handle response status codes from GET orgs/current endpoint', () => {
  const orgId = 'test-id'

  test('200 status', async () => {
    const result = await getOrg({ orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('name')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getOrg({ orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
