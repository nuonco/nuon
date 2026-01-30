import '@test/mock-auth'
import { describe, expect, test } from 'vitest'
import { forgetComponent } from './forget-component'

describe('forgetComponent should handle response status codes from POST installs/:installId/components/:componentId/forget endpoint', () => {
  const componentId = 'test-component-id'
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status with boolean response', async () => {
    const { data } = await forgetComponent({
      componentId,
      installId,
      orgId,
    })
    expect(data).toBeDefined()
    expect(typeof data).toBe('boolean')
  })

  test.each([400, 404, 500])('%s status', async (code) => {
    const { error, status } = await forgetComponent({
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
