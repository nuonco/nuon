import { describe, expect, test } from 'vitest'
import {
  deprovisionSandbox,
  type TDeprovisionSandboxBody,
} from './deprovision-sandbox'

describe('deprovisionSandbox should handle response status codes from POST installs/:id/deprovision-sandbox endpoint', () => {
  const orgId = 'test-org-id'
  const installId = 'test-install-id'

  test('201 status with plan_only: true', async () => {
    const body: TDeprovisionSandboxBody = { plan_only: true }
    const result = await deprovisionSandbox({ body, installId, orgId })
    expect(result).toHaveProperty('data')
  })
})
