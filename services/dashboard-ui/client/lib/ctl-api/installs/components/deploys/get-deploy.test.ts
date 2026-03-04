import { badResponseCodes } from '@test/utils'
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

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getDeploy({ installId, deployId, orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
