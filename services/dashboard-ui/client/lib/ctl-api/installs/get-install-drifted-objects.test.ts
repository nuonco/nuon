import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallDriftedObjects } from './get-install-drifted-objects'

describe('getInstallDriftedObjects should handle response status codes from GET installs/:installId/drifted-objects endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getInstallDriftedObjects({ installId, orgId })
    expect(Array.isArray(result)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      getInstallDriftedObjects({ installId, orgId })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
