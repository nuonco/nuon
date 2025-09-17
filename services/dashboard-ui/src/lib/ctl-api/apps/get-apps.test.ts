import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getApps } from './get-apps'

describe('getApps should handle response status codes from GET apps endpoint', () => {
  const orgId = 'test-id'
  test('200 status with all optional params', async () => {
    const { data: spec } = await getApps({
      orgId,
      q: 'test-query',
      limit: 10,
      offset: 0,
    })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('name')
    })
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getApps({ orgId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
