import { describe, expect, test } from 'vitest'
import { syncSecrets, type TSyncSecretsBody } from './sync-secrets'

describe('syncSecrets should handle response status codes from POST installs/:id/sync-secrets endpoint', () => {
  const orgId = 'test-org-id'
  const installId = 'test-install-id'

  test('201 status with plan_only: true', async () => {
    const body: TSyncSecretsBody = { plan_only: true }
    const result = await syncSecrets({ body, installId, orgId })
    expect(result).toHaveProperty('data')
  })
})
