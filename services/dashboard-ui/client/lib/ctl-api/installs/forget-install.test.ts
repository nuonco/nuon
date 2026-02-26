import '@test/mock-auth'
import { describe, expect, test } from 'vitest'
import { forgetInstall } from './forget-install'

describe('forgetInstall should handle response status codes from POST installs/:id/forget endpoint', () => {
  const orgId = 'test-org-id'
  const installId = 'test-install-id'

  test('200 status', async () => {
    const { data } = await forgetInstall({
      installId,
      orgId,
    })
    expect(data).toBe(true)
  })

  test.each([400, 404, 500])('%s status', async (code) => {
    const { error, status } = await forgetInstall({
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
