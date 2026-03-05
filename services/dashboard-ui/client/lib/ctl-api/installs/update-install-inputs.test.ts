import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { updateInstallInputs } from './update-install-inputs'

describe('updateInstallInputs should handle response status codes from PATCH installs/:installId/inputs endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status with inputs', async () => {
    const result = await updateInstallInputs({
      installId,
      orgId,
      body: { inputs: { 'input-key-1': 'input-value-1', 'input-key-2': 'input-value-2' } },
    })
    expect(result).toHaveProperty('values')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      updateInstallInputs({
        installId,
        orgId,
        body: { inputs: { 'test-key': 'test-value' } },
      })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
