import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import {
  createInstallConfig,
  type TCreateInstallConfigBody,
} from './create-install-config'

describe('createInstallConfig should handle response status codes from POST installs/:id/config endpoint', () => {
  const orgId = 'test-org-id'
  const installId = 'test-install-id'

  test('201 status with approve-all option', async () => {
    const body: TCreateInstallConfigBody = {
      approval_option: 'approve-all',
    }

    const { data, status } = await createInstallConfig({
      body,
      installId,
      orgId,
    })
    expect(data).toHaveProperty('id')
    expect(data).toHaveProperty('approval_option')
    expect(status).toBe(201)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const body: TCreateInstallConfigBody = {
      approval_option: 'approve-all',
    }

    const { error, status } = await createInstallConfig({
      body,
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
