import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getLogStreamById } from './get-log-stream-by-id'

describe('getLogStreamById should handle response status codes from GET log-streams/:logStreamId endpoint', () => {
  const logStreamId = 'test-log-stream-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const { data: logStream } = await getLogStreamById({
      logStreamId,
      orgId,
    })
    expect(logStream).toHaveProperty('id')
    expect(logStream).toHaveProperty('owner_type')
  }, 60000)

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getLogStreamById({
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
