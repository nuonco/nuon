import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getVCSConnectionRepos } from './get-vcs-connection-repos'

describe('getVCSConnectionRepos should handle response status codes', () => {
  const orgId = 'test-org-id'
  const connectionId = 'test-connection-id'

  test('200 status', async () => {
    const { data } = await getVCSConnectionRepos({ orgId, connectionId })
    expect(data).toHaveProperty('repositories')
    expect(data).toHaveProperty('total_count')
    expect(Array.isArray(data.repositories)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getVCSConnectionRepos({
      orgId,
      connectionId,
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
