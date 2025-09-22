import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getAppConfigs } from './get-app-configs'

describe('getAppConfigs should handle response status codes from GET app configs endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-app-id'

  test('200 status', async () => {
    const { data: configs } = await getAppConfigs({ orgId, appId })
    configs.forEach((config) => {
      expect(config).toHaveProperty('id')
      expect(config).toHaveProperty('status')
    })
  }, 30000)

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getAppConfigs({ orgId, appId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
