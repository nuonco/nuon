import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import {
  deprovisionInstall,
  type TDeprovisionInstallBody,
} from './deprovision-install'

describe('deprovisionInstall should handle response status codes from POST installs/:id/deprovision endpoint', () => {
  const orgId = 'test-org-id'
  const installId = 'test-install-id'

  test('201 status with plan_only: true', async () => {
    const body: TDeprovisionInstallBody = {
      plan_only: true,
    }

    const { data, status } = await deprovisionInstall({
      body,
      installId,
      orgId,
    })
    expect(typeof data).toBe('string')
    expect(status).toBe(201)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const body: TDeprovisionInstallBody = {
      plan_only: true,
    }

    const { error, status } = await deprovisionInstall({
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
