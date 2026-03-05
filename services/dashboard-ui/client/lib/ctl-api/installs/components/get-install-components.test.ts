import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallComponents } from './get-install-components'

describe('getInstallComponents should handle response status codes from GET installs/:installId/components endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status with pagination', async () => {
    const result = await getInstallComponents({ installId, limit: 10, orgId, offset: 0 })
    expect(Array.isArray(result)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      getInstallComponents({ installId, orgId })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
