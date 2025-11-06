import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getLogsByLogStreamId } from './get-logs-by-log-stream-id'

describe('getLogsByLogStreamId should handle response status codes from GET log-streams/:logStreamId/logs endpoint', () => {
  const logStreamId = 'test-log-stream-id'
  const orgId = 'test-org-id'

  test('200 status with offset', async () => {
    const { data: logs } = await getLogsByLogStreamId({
      logStreamId,
      orgId,
      offset: 'some-offset-token',
    })
    expect(logs).toBeInstanceOf(Array)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getLogsByLogStreamId({
      logStreamId,
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
