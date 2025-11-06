import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { shutdownRunnerInstance } from './shutdown-runner-instance'

describe('shutdownRunnerInstance should handle response status codes from POST runners/:id/mng/shutdown-vm endpoint', () => {
  const orgId = 'test-org-id'
  const runnerId = 'test-runner-id'

  test('200 status', async () => {
    const { data } = await shutdownRunnerInstance({
      orgId,
      runnerId,
    })
    expect(data).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await shutdownRunnerInstance({
      orgId,
      runnerId,
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
