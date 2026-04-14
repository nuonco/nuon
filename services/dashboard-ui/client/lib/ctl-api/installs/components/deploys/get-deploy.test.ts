import { describe, expect, test } from 'vitest'
import { getDeploy } from './get-deploy'

describe('getDeploy should handle response status codes from GET installs/:id/deploys/:deployId endpoint', () => {
  const installId = 'test-install-id'
  const deployId = 'test-deploy-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getDeploy({ installId, deployId, orgId })
    expect(result).toHaveProperty('id')
  })
})
