import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getVCSConnectionRepos } from './get-vcs-connection-repos'

describe('getVCSConnectionRepos should handle response status codes', () => {
  const orgId = 'test-org-id'
  const connectionId = 'test-connection-id'

  test('200 status', async () => {
    const result = await getVCSConnectionRepos({ orgId, connectionId })
    expect(result).toHaveProperty('repositories')
    expect(result).toHaveProperty('total_count')
    expect(Array.isArray(result.repositories)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      getVCSConnectionRepos({ orgId, connectionId })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
