import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallSandboxRuns } from './get-install-sandbox-runs'

describe('getInstallSandboxRuns should handle response status codes from GET installs/:installId/sandbox-runs endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status with pagination', async () => {
    const { data: runs, status } = await getInstallSandboxRuns({
      installId,
      orgId,
      limit: 10,
      offset: 0,
    })
    expect(status).toBe(200)
    expect(runs).toBeInstanceOf(Array)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getInstallSandboxRuns({
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
