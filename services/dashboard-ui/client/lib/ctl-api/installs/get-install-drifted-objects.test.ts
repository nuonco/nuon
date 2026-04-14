import { describe, expect, test } from 'vitest'
import { getInstallDriftedObjects } from './get-install-drifted-objects'

describe('getInstallDriftedObjects should handle response status codes from GET installs/:installId/drifted-objects endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getInstallDriftedObjects({ installId, orgId })
    expect(Array.isArray(result)).toBe(true)
  })
})
