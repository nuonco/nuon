import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { buildComponents } from './build-components'

describe('buildComponents should handle response status codes from POST apps/:appId/components/build-all endpoint', () => {
  const appId = 'test-app-id'
  const orgId = 'test-org-id'

  test('200 status with default body', async () => {
    const result = await buildComponents({ appId, orgId })
    expect(Array.isArray(result)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(buildComponents({ appId, orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
