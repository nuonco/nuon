import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getRunnerLatestHeartbeat } from './get-runner-latest-heartbeat'

describe('getRunnerLatestHeartbeat should handle response status codes from GET runners/:id/heart-beats/latest endpoint', () => {
  const runnerId = 'test-id'
  const orgId = 'test-id'

  test('200 status', async () => {
    const result = await getRunnerLatestHeartbeat({ runnerId, orgId })
    expect(result).toBeDefined()
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      getRunnerLatestHeartbeat({ runnerId, orgId })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
