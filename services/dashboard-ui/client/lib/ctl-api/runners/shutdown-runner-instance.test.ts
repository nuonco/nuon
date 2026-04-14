import { describe, expect, test } from 'vitest'
import { shutdownRunnerInstance } from './shutdown-runner-instance'

describe('shutdownRunnerInstance should handle response status codes from POST runners/:id/mng/shutdown-vm endpoint', () => {
  const orgId = 'test-org-id'
  const runnerId = 'test-runner-id'

  test('200 status', async () => {
    const result = await shutdownRunnerInstance({ orgId, runnerId })
    expect(result).toBe(true)
  })
})
