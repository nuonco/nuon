import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallComponents } from './get-install-components'

describe('getInstallComponents should handle response status codes from GET installs/:installId/components endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status with pagination', async () => {
    const { data: deploys } = await getInstallComponents({
      installId,
      limit: 10,
      orgId,
      offset: 0,
    })
    expect(deploys).toBeInstanceOf(Array)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getInstallComponents({
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
