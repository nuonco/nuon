import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getAppConfigById } from './get-app-config-by-id'

describe('getAppConfigById should handle response status codes from GET app config by id endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-app-id'
  const appConfigId = 'test-app-config-id'

  test('200 status', async () => {
    const { data: config } = await getAppConfigById({
      orgId,
      appId,
      appConfigId,
    })
    expect(config).toHaveProperty('id')
    expect(config).toHaveProperty('status')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getAppConfigById({
      orgId,
      appId,
      appConfigId,
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
