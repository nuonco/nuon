import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getAppInstalls } from './get-app-installs'

describe('getAppInstalls should handle response status codes from GET /apps/:appId/installs?:params endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-app-id'

  test('200 status', async () => {
    const { data: spec } = await getAppInstalls({
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
    const { error, status } = await getAppInstalls({ appId, orgId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
