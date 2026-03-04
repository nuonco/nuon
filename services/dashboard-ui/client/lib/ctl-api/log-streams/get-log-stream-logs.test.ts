import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getLogStreamLogs } from './get-log-stream-logs'

describe('getLogStreamLogs should handle response status codes from GET log-streams/:logStreamId/logs endpoint', () => {
  const logStreamId = 'test-log-stream-id'
  const orgId = 'test-org-id'

  test('200 status with offset', async () => {
    const result = await getLogStreamLogs({ logStreamId, orgId, offset: 'some-offset-token' })
    expect(Array.isArray(result)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      getLogStreamLogs({ logStreamId, orgId })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
