import { describe, expect, test } from 'vitest'
import { restartMngRunner } from './restart-mng-runner'

describe('restartMngRunner should handle response status codes from POST runners/:id/mng/restart endpoint', () => {
  const orgId = 'test-org-id'
  const runnerId = 'test-runner-id'

  test('200 status', async () => {
    const result = await restartMngRunner({ orgId, runnerId })
    expect(result).toBe(true)
  })
})
