import { describe, expect, test } from 'vitest'
import { getInstallSandboxRuns } from './get-install-sandbox-runs'

describe('getInstallSandboxRuns should handle response status codes from GET installs/:installId/sandbox-runs endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status with pagination', async () => {
    const result = await getInstallSandboxRuns({ installId, orgId, limit: 10, offset: 0 })
    expect(Array.isArray(result.data)).toBe(true)
  })
})
