import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { buildComponent } from './build-component'

describe('buildComponent should handle response status codes from POST components/:componentId/build endpoint', () => {
  const componentId = 'test-component-id'
  const orgId = 'test-org-id'

  test('200 status with default body', async () => {
    const result = await buildComponent({ componentId, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('status_v2')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(buildComponent({ componentId, orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
