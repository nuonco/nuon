import { describe, expect, test } from 'vitest'
import { getVCSConnection } from './get-vcs-connection'

describe('getVCSConnection should handle response status codes from GET vcs/connections/:id endpoint', () => {
  const orgId = 'test-org-id'
  const connectionId = 'test-connection-id'

  test('200 status', async () => {
    const result = await getVCSConnection({ orgId, connectionId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('github_install_id')
  })
})
