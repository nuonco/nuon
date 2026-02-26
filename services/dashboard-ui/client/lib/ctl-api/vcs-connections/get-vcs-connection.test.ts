import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getVCSConnection } from './get-vcs-connection'

describe('getVCSConnection should handle response status codes from GET vcs/connections/:id endpoint', () => {
  const orgId = 'test-org-id'
  const connectionId = 'test-connection-id'

  test('200 status', async () => {
    const { data } = await getVCSConnection({ orgId, connectionId })
    expect(data).toHaveProperty('id')
    expect(data).toHaveProperty('github_install_id')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getVCSConnection({
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
