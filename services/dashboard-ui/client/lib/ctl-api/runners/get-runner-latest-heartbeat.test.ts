import { describe, expect, test } from 'vitest'
import { getRunnerLatestHeartbeat } from './get-runner-latest-heartbeat'

describe('getRunnerLatestHeartbeat should handle response status codes from GET runners/:id/heart-beats/latest endpoint', () => {
  const runnerId = 'test-id'
  const orgId = 'test-id'

  test('200 status', async () => {
    const result = await getRunnerLatestHeartbeat({ runnerId, orgId })
    expect(result).toBeDefined()
  })
})
