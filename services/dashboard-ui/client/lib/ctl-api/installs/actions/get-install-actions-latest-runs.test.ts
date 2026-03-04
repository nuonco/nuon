import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallActionsLatestRuns } from './get-install-actions-latest-runs'

describe('getInstallActionsLatestRuns should handle response status codes from GET installs/{installId}/action-workflows/latest-runs endpoint', () => {
  const orgId = 'test-org-id'
  const installId = 'test-install-id'

  test('200 status with all optional params', async () => {
    const result = await getInstallActionsLatestRuns({
      installId,
      orgId,
      q: 'test-query',
      limit: 10,
      offset: 0,
    })
    expect(Array.isArray(result)).toBe(true)
  }, 60000)

  test.each(badResponseCodes)(
    '%s status',
    async () => {
      await expect(
        getInstallActionsLatestRuns({ installId, orgId })
      ).rejects.toMatchObject({
        error: expect.any(String),
        description: expect.any(String),
        user_error: expect.any(Boolean),
      })
    },
    30000
  )
})
