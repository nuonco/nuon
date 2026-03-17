import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getPendingApprovals } from './get-pending-approvals'

describe('getPendingApprovals should handle response status codes from GET workflows/pending-approvals endpoint', () => {
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getPendingApprovals({ orgId })
    expect(result).toBeInstanceOf(Array)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getPendingApprovals({ orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
