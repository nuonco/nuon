import { badResponseCodes } from '@test/utils'
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

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      getAppConfigGraph({ orgId, appId, appConfigId })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
