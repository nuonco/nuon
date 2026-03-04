import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallCurrentInputs } from './get-install-current-inputs'

describe('getInstallCurrentInputs should handle response status codes from GET installs/:installId/inputs/current endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getInstallCurrentInputs({ installId, orgId })
    expect(result).toHaveProperty('values')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      getInstallCurrentInputs({ installId, orgId })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
