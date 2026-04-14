import { describe, expect, test } from 'vitest'
import { getComponentDeploys } from './get-component-deploys'

describe('getComponentDeploys should handle response status codes from GET installs/:installId/components/:componentId/deploys endpoint', () => {
  const installId = 'test-install-id'
  const componentId = 'test-component-id'
  const orgId = 'test-org-id'

  test('200 status with pagination', async () => {
    const result = await getComponentDeploys({
      installId,
      componentId,
      orgId,
      limit: 10,
      offset: 0,
    })
    expect(Array.isArray(result.data)).toBe(true)
  }, 60000)
})
