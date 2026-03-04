import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallState } from './get-install-state'

describe('getInstallState should handle response status codes from GET installs/:id/state endpoint', () => {
  const installId = 'test-id'
  const orgId = 'test-id'

  test('200 status', async () => {
    const result = await getInstallState({ installId, orgId })
    expect(result).toBeDefined()
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getInstallState({ installId, orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
