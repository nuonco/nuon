import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getRunnerLatestHeartbeat } from './get-runner-latest-heartbeat'

describe('getRunnerLatestHeartbeat should handle response status codes from GET runners/:id/heart-beats/latest endpoint', () => {
  const runnerId = 'test-id'
  const orgId = 'test-id'
  test('200 status', async () => {
    const { data: spec } = await getRunnerLatestHeartbeat({ runnerId, orgId })
    Object.values(spec).forEach((hb) => {
      expect(hb).toHaveProperty('process')
      expect(hb).toHaveProperty('alive_time')
      expect(hb).toHaveProperty('created_at')
    })
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getRunnerLatestHeartbeat({
      runnerId,
      orgId,
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
