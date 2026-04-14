import { describe, expect, test } from 'vitest'
import { getPendingApprovals } from './get-pending-approvals'

describe('getPendingApprovals should handle response status codes from GET workflows/pending-approvals endpoint', () => {
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getPendingApprovals({ orgId })
    expect(result).toBeInstanceOf(Array)
  })
})
