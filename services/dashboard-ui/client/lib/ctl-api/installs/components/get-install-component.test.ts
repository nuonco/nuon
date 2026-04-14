import { describe, expect, test } from 'vitest'
import { getInstallComponent } from './get-install-component'

describe('getInstallComponent should handle response status codes from GET installs/:installId/components/:componentId endpoint', () => {
  const installId = 'test-install-id'
  const componentId = 'test-component-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getInstallComponent({ installId, componentId, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('component_id')
  }, 60000)
})
