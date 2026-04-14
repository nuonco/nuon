import { describe, expect, test } from 'vitest'
import { getAppConfigs } from './get-app-configs'

describe('getAppConfigs should handle response status codes from GET app configs endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-app-id'

  test('200 status', async () => {
    const result = await getAppConfigs({ orgId, appId })
    expect(Array.isArray(result)).toBe(true)
  }, 30000)
})
