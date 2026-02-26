import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getLogStream } from './get-log-stream'

describe('getLogStream should handle response status codes from GET log-streams/:logStreamId endpoint', () => {
  const logStreamId = 'test-log-stream-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const { data: logStream } = await getLogStream({
      logStreamId,
      orgId,
    })
    expect(logStream).toHaveProperty('id')
    expect(logStream).toHaveProperty('owner_type')
  }, 60000)

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getLogStream({
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
