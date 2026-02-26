import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { updateInstallInputs } from './update-install-inputs'

describe('updateInstallInputs should handle response status codes from PATCH installs/:installId/inputs endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status with inputs', async () => {
    const { data: installInputs } = await updateInstallInputs({
      installId,
      orgId,
      body: {
        inputs: {
          'input-key-1': 'input-value-1',
          'input-key-2': 'input-value-2',
        },
      },
    })
    expect(installInputs).toHaveProperty('values')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await updateInstallInputs({
      installId,
      orgId,
      body: {
        inputs: {
          'test-key': 'test-value',
        },
      },
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
