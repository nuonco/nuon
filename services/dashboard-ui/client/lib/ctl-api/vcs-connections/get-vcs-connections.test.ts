import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getVCSConnections } from './get-vcs-connections'

describe('getVCSConnections should handle response status codes from GET vcs/connections endpoint', () => {
  const orgId = 'test-id'

  test('200 status', async () => {
    const result = await getVCSConnections({ orgId })
    expect(Array.isArray(result)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getVCSConnections({ orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
