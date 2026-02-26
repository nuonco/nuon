import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getApp } from './get-app'

describe('getApp should handle response status codes from GET apps/:id endpoint', () => {
  const appId = 'test-id'
  const orgId = 'test-id'
  test('200 status', async () => {
    const { data: spec } = await getApp({ appId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('name')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getApp({ appId, orgId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
