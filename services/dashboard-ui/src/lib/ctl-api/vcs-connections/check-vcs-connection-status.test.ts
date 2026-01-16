import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { checkVCSConnectionStatus } from './check-vcs-connection-status'

describe('checkVCSConnectionStatus should handle response status codes from GET vcs/connections/:id/check-status endpoint', () => {
  const orgId = 'test-org-id'
  const connectionId = 'test-connection-id'

  test('200 status', async () => {
    const { data } = await checkVCSConnectionStatus({ orgId, connectionId })
    expect(data).toHaveProperty('status')
    expect(data).toHaveProperty('github_install_id')
    expect(data).toHaveProperty('checked_at')
    expect(data).toHaveProperty('account')
    expect(data).toHaveProperty('permissions')
    expect(data).toHaveProperty('repository_selection')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await checkVCSConnectionStatus({
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
