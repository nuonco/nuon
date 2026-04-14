import { describe, expect, test } from 'vitest'
import { getInstallWorkflows } from './get-install-workflows'

describe('getInstallWorkflows should handle response status codes from GET installs/:installId/workflows endpoint', () => {
  const orgId = 'test-id'
  const installId = 'test-install-id'

  test('200 status with all optional params', async () => {
    const result = await getInstallWorkflows({ orgId, installId, limit: 10, offset: 0 })
    expect(Array.isArray(result.data)).toBe(true)
  }, 30000)
})
