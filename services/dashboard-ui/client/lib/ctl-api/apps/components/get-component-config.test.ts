import { describe, expect, test } from 'vitest'
import { getComponentConfig } from './get-component-config'

describe('getComponentConfig should handle response status codes from GET apps/:appId/components/:componentId/configs/:configId endpoint', () => {
  const appId = 'test-app-id'
  const componentId = 'test-component-id'
  const configId = 'test-config-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getComponentConfig({ appId, componentId, configId, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('component_id')
  }, 60000)
})
