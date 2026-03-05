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
    const body: TDeprovisionInstallBody = { plan_only: true }
    const result = await deprovisionInstall({ body, installId, orgId })
    expect(result).toHaveProperty('data')
  })

  test.each(badResponseCodes)('%s status', async () => {
    const body: TDeprovisionInstallBody = { plan_only: true }
    await expect(
      deprovisionInstall({ body, installId, orgId })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
