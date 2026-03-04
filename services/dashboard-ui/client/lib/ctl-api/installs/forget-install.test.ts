import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { forgetInstall } from './forget-install'

describe('forgetInstall should handle response status codes from POST installs/:id/forget endpoint', () => {
  const orgId = 'test-org-id'
  const installId = 'test-install-id'

  test('200 status', async () => {
    const result = await forgetInstall({ installId, orgId })
    expect(result).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(forgetInstall({ installId, orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
