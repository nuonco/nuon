import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallDriftedObjects } from './get-install-drifted-objects'

describe('getInstallDriftedObjects should handle response status codes from GET installs/:installId/drifted-objects endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const { data: driftedObjects } = await getInstallDriftedObjects({
      installId,
      orgId,
    })
    expect(Array.isArray(driftedObjects)).toBe(true)
    if (driftedObjects && driftedObjects.length > 0) {
      expect(driftedObjects[0]).toHaveProperty('target_id')
      expect(driftedObjects[0]).toHaveProperty('target_type')
    }
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getInstallDriftedObjects({
      installId,
      orgId,
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
