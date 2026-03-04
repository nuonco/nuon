import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { shutdownMngRunner } from './shutdown-mng-runner'

describe('shutdownMngRunner should handle response status codes from POST runners/:id/mng/shutdown endpoint', () => {
  const orgId = 'test-org-id'
  const runnerId = 'test-runner-id'

  test('200 status', async () => {
    const result = await shutdownMngRunner({ orgId, runnerId })
    expect(result).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(shutdownMngRunner({ orgId, runnerId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
