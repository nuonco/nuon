import { describe, expect, test } from 'vitest'
import { deployComponent } from './deploy-component'

describe('deployComponent should handle response status codes from POST installs/:installId/deploys endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status with default body', async () => {
    const result = await deployComponent({
      body: { build_id: 'test-build-id', deploy_dependents: true, plan_only: true },
      installId,
      orgId,
    })
    expect(result).toHaveProperty('data')
    expect(result.data).toHaveProperty('id')
    expect(result.data).toHaveProperty('status_v2')
  })
})
