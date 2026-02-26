import '@test/mock-auth'
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
    const body: TReprovisionSandboxBody = {
      plan_only: true,
    }

    const { data, status } = await reprovisionSandbox({
      body,
      installId,
      orgId,
    })
    expect(typeof data).toBe('string')
    expect(status).toBe(201)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const body: TReprovisionSandboxBody = {
      plan_only: true,
    }

    const { error, status } = await reprovisionSandbox({
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
