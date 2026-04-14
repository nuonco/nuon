import { describe, expect, test } from 'vitest'
import {
  reprovisionInstall,
  type TReprovisionInstallBody,
} from './reprovision-install'

describe('reprovisionInstall should handle response status codes from POST installs/:id/reprovision endpoint', () => {
  const orgId = 'test-org-id'
  const installId = 'test-install-id'

  test('201 status with plan_only: true', async () => {
    const body: TReprovisionInstallBody = { plan_only: true }
    const result = await reprovisionInstall({ body, installId, orgId })
    expect(result).toHaveProperty('data')
  })
})
