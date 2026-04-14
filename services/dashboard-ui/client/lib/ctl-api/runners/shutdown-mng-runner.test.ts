import { describe, expect, test } from 'vitest'
import { shutdownMngRunner } from './shutdown-mng-runner'

describe('shutdownMngRunner should handle response status codes from POST runners/:id/mng/shutdown endpoint', () => {
  const orgId = 'test-org-id'
  const runnerId = 'test-runner-id'

  test('200 status', async () => {
    const result = await shutdownMngRunner({ orgId, runnerId })
    expect(result).toBe(true)
  })
})
