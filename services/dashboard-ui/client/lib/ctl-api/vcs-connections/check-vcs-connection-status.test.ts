import { describe, expect, test } from 'vitest'
import { checkVCSConnectionStatus } from './check-vcs-connection-status'

describe('checkVCSConnectionStatus should handle response status codes from GET vcs/connections/:id/check-status endpoint', () => {
  const orgId = 'test-org-id'
  const connectionId = 'test-connection-id'

  test('200 status', async () => {
    const result = await checkVCSConnectionStatus({ orgId, connectionId })
    expect(result).toHaveProperty('status')
    expect(result).toHaveProperty('github_install_id')
    expect(result).toHaveProperty('checked_at')
    expect(result).toHaveProperty('account')
    expect(result).toHaveProperty('permissions')
    expect(result).toHaveProperty('repository_selection')
  })
})
