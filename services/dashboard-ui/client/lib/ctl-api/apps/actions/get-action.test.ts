import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getAction } from './get-action'

describe('getAction should handle response status codes from GET apps/:appId/action-workflows/:actionId endpoint', () => {
  const actionId = 'test-action-id'
  const appId = 'test-app-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getAction({ actionId, appId, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('status')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getAction({ actionId, appId, orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
