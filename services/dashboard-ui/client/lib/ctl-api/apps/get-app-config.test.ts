import { describe, expect, test } from 'vitest'
import { getAppConfig } from './get-app-config'

describe('getAppConfig should handle response status codes from GET app config by id endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-app-id'
  const appConfigId = 'test-app-config-id'

  test('200 status', async () => {
    const result = await getAppConfig({ orgId, appId, appConfigId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('status')
  })
})
