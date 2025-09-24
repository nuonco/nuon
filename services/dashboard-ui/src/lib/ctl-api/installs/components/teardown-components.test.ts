import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { teardownComponents } from './teardown-components'

describe('teardownComponents should handle response status codes from POST installs/:installId/components/teardown-all endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status with default body', async () => {
    const { data: deploy } = await teardownComponents({
      body: {
        plan_only: true,
      },
      installId,
      orgId,
    })
    expect(deploy).toBeDefined()
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await teardownComponents({
      body: {
        plan_only: true,
      },
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
