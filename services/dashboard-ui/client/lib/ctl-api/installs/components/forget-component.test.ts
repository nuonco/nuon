import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { forgetComponent } from './forget-component'

describe('forgetComponent should handle response status codes from POST installs/:installId/components/:componentId/forget endpoint', () => {
  const componentId = 'test-component-id'
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status with boolean response', async () => {
    const result = await forgetComponent({ componentId, installId, orgId })
    expect(result).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      forgetComponent({ componentId, installId, orgId })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
