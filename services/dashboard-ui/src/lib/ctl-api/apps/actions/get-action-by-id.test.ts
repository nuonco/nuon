import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getActionById } from './get-action-by-id'

describe('getActionById should handle response status codes from GET apps/:appId/action-workflows/:actionId endpoint', () => {
  const actionId = 'test-action-id'
  const appId = 'test-app-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const { data: build } = await getActionById({
      actionId,
      appId,
      orgId,
    })
    expect(build).toHaveProperty('id')
    expect(build).toHaveProperty('status')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getActionById({
      actionId,
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
