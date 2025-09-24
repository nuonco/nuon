import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { teardownComponent } from './teardown-component'

describe('teardownComponent should handle response status codes from POST installs/:installId/components/:componentId/teardown endpoint', () => {
  const componentId = 'test-component-id'
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status with default body', async () => {
    const { data: deploy } = await teardownComponent({
      body: {
        error_behavior: 'continue',
        plan_only: true,
      },
      componentId,
      installId,
      orgId,
    })
    expect(deploy).toBeDefined()
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await teardownComponent({
      body: {
        error_behavior: 'continue',
        plan_only: true,
      },
      componentId,
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
