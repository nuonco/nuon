import { describe, expect, test } from 'vitest'
import { getInstallAction } from './get-install-action'

describe('getInstallAction should handle response status codes from GET installs/:installId/action-workflows/:actionId/recent-runs endpoint', () => {
  const installId = 'test-install-id'
  const actionId = 'test-action-id'
  const orgId = 'test-org-id'

  test('200 status with pagination', async () => {
    const result = await getInstallAction({
      installId,
      actionId,
      orgId,
      limit: 10,
      offset: 0,
    })
    expect(result).toHaveProperty('id')
  })
})
