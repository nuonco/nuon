import { describe, expect, test } from 'vitest'
import { createVCSConnection } from './create-vcs-connection'

describe('createVCSConnection should handle response status codes from POST vcs/connections endpoint', () => {
  const orgId = 'test-org-id'
  const body = { github_install_id: 'test-github-install-id' }

  test('201 status', async () => {
    const result = await createVCSConnection({ orgId, body })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('github_install_id')
  })
})
