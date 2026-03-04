import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getComponent } from './get-component'

describe('getComponent should handle response status codes from GET components/:componentId endpoint', () => {
  const componentId = 'test-component-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getComponent({ componentId, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('status')
  }, 60000)

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getComponent({ componentId, orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
