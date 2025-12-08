import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { buildComponents } from './build-components'

describe('buildComponents should handle response status codes from POST apps/:appId/components/build-all endpoint', () => {
  const appId = 'test-app-id'
  const orgId = 'test-org-id'

  test('200 status with default body', async () => {
    const { data: builds } = await buildComponents({
      appId,
      orgId,
    })
    expect(Array.isArray(builds)).toBe(true)
    if (builds && builds.length > 0) {
      expect(builds[0]).toHaveProperty('id')
      expect(builds[0]).toHaveProperty('status_v2')
    }
  })


  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await buildComponents({
      appId,
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
