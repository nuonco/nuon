import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import {
  reprovisionSandbox,
  type TReprovisionSandboxBody,
} from './reprovision-sandbox'

describe('reprovisionSandbox should handle response status codes from POST installs/:id/reprovision-sandbox endpoint', () => {
  const orgId = 'test-org-id'
  const installId = 'test-install-id'

  test('201 status with plan_only: true', async () => {
    const body: TReprovisionSandboxBody = { plan_only: true }
    const result = await reprovisionSandbox({ body, installId, orgId })
    expect(result).toHaveProperty('data')
  })

  test.each(badResponseCodes)('%s status', async () => {
    const body: TReprovisionSandboxBody = { plan_only: true }
    await expect(
      reprovisionSandbox({ body, installId, orgId })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
