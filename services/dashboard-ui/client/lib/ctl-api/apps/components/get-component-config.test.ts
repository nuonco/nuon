import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getComponentConfig } from './get-component-config'

describe('getComponentConfig should handle response status codes from GET apps/:appId/components/:componentId/configs/:configId endpoint', () => {
  const appId = 'test-app-id'
  const componentId = 'test-component-id'
  const configId = 'test-config-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const { data: config } = await getComponentConfig({
      appId,
      componentId,
      configId,
      orgId,
    })
    expect(config).toHaveProperty('id')
    expect(config).toHaveProperty('component_id')
  }, 60000)

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getComponentConfig({
      appId,
      componentId,
      configId,
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