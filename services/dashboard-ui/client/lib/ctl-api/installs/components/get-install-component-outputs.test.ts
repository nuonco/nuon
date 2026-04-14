import { describe, expect, test } from 'vitest'
import { getInstallComponentOutputs } from './get-install-component-outputs'

describe('getInstallComponentOutputs should handle response status codes from GET installs/:installId/components/:componentId/outputs endpoint', () => {
  const componentId = 'test-id'
  const installId = 'test-id'
  const orgId = 'test-id'

  test('200 status', async () => {
    const result = await getInstallComponentOutputs({ componentId, installId, orgId })
    expect(result).toBeDefined()
  })
})
