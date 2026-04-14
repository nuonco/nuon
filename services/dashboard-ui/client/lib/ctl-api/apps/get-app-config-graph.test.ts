import { describe, expect, test } from 'vitest'
import { getAppConfigGraph } from './get-app-config-graph'

describe('getAppConfigGraph should handle response status codes from GET apps/:appId/configs/:configId/graph endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-app-id'
  const appConfigId = 'test-app-config-id'

  test('200 status', async () => {
    const result = await getAppConfigGraph({ orgId, appId, appConfigId })
    expect(result).toBeDefined()
  })
})
