import { describe, expect, test } from 'vitest'
import { getInstallStack } from './get-install-stack'

describe('getInstallStack should handle response status codes from GET installs/:id/stack endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getInstallStack({ installId, orgId })
    expect(result).toBeDefined()
    expect(result).toHaveProperty('id')
  })
})
