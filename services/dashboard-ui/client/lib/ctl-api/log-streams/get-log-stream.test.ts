import { describe, expect, test } from 'vitest'
import { getLogStream } from './get-log-stream'

describe('getLogStream should handle response status codes from GET log-streams/:logStreamId endpoint', () => {
  const logStreamId = 'test-log-stream-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getLogStream({ logStreamId, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('owner_type')
  }, 60000)
})
