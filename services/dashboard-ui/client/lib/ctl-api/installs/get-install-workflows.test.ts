import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallWorkflows } from './get-install-workflows'

describe('getInstallWorkflows should handle response status codes from GET installs/:installId/workflows endpoint', () => {
  const orgId = 'test-id'
  const installId = 'test-install-id'

  test('200 status with all optional params', async () => {
    const { data: workflows } = await getInstallWorkflows({
      orgId,
      installId,
      limit: 10,
      offset: 0,
    })
    workflows.forEach((workflow) => {
      expect(workflow).toHaveProperty('id')
      expect(workflow).toHaveProperty('name')
    })
  }, 30000)

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getInstallWorkflows({ orgId, installId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
