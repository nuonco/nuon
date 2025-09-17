import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallsByAppId } from './get-installs-by-app-id'

describe('getInstallsByAppId should handle response status codes from GET /apps/:appId/installs?:params endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-app-id'

  test('200 status', async () => {
    const { data: spec } = await getInstallsByAppId({
      appId,
      orgId,
      limit: 10,
      offset: 0,
    })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('name')
    })
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getInstallsByAppId({ appId, orgId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
