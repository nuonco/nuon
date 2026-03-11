import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { restartMngRunner } from './restart-mng-runner'

describe('restartMngRunner should handle response status codes from POST runners/:id/mng/restart endpoint', () => {
  const orgId = 'test-org-id'
  const runnerId = 'test-runner-id'

  test('200 status', async () => {
    const result = await restartMngRunner({ orgId, runnerId })
    expect(result).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(restartMngRunner({ orgId, runnerId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
