import { describe, expect, test } from 'vitest'
import { getInstallState } from './get-install-state'

describe('getInstallState should handle response status codes from GET installs/:id/state endpoint', () => {
  const installId = 'test-id'
  const orgId = 'test-id'

  test('200 status', async () => {
    const result = await getInstallState({ installId, orgId })
    expect(result).toBeDefined()
  })
})
