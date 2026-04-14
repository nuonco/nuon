import { describe, expect, test } from 'vitest'
import { updateMngRunner } from './update-mng-runner'

describe('updateMngRunner should handle response status codes from POST runners/:id/mng/update endpoint', () => {
  const orgId = 'test-org-id'
  const runnerId = 'test-runner-id'

  test('200 status', async () => {
    const result = await updateMngRunner({ orgId, runnerId })
    expect(result).toBe(true)
  })
})
