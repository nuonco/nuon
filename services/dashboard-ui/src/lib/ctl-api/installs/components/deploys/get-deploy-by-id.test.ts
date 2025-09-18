import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getDeployById } from './get-deploy-by-id'

describe('getDeployById should handle response status codes from GET installs/:id/deploys/:deployId endpoint', () => {
  const installId = 'test-install-id'
  const deployId = 'test-deploy-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const { data: deploy, status } = await getDeployById({
      installId,
      deployId,
      orgId,
    })
    expect(status).toBe(200)
    expect(deploy).toHaveProperty('id')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getDeployById({
      installId,
      deployId,
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
