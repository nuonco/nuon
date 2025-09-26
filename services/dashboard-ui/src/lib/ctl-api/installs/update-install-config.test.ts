import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import {
  updateInstallConfig,
  type TUpdateInstallConfigBody,
} from './update-install-config'

describe('updateInstallConfig should handle response status codes from PATCH installs/:id/configs/:configId endpoint', () => {
  const orgId = 'test-org-id'
  const installId = 'test-install-id'
  const installConfigId = 'test-install-config-id'

  test('201 status with approve-all option', async () => {
    const body: TUpdateInstallConfigBody = {
      approval_option: 'approve-all',
    }

    const { data, status } = await updateInstallConfig({
      body,
      installConfigId,
      installId,
      orgId,
    })
    expect(data).toHaveProperty('id')
    expect(data).toHaveProperty('approval_option')
    expect(status).toBe(201)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const body: TUpdateInstallConfigBody = {
      approval_option: 'approve-all',
    }

    const { error, status } = await updateInstallConfig({
      body,
      installConfigId,
      installId,
      orgId,
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
