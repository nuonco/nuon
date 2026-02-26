import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { shutdownMngRunner } from './shutdown-mng-runner'

describe('shutdownMngRunner should handle response status codes from POST runners/:id/mng/shutdown endpoint', () => {
  const orgId = 'test-org-id'
  const runnerId = 'test-runner-id'

  test('200 status', async () => {
    const { data } = await shutdownMngRunner({
      orgId,
      runnerId,
    })
    expect(data).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await shutdownMngRunner({
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