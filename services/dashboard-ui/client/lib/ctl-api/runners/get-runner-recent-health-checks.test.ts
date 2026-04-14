import { describe, expect, test } from 'vitest'
import { getRunnerRecentHealthChecks } from './get-runner-recent-health-checks'

describe('getRunnerRecentHealthChecks should handle response status codes from GET runners/:id/recent-health-checks endpoint', () => {
  const runnerId = 'test-id'
  const orgId = 'test-id'

  test('200 status with all parameters', async () => {
    const result = await getRunnerRecentHealthChecks({
      runnerId,
      orgId,
      limit: 10,
      offset: 0,
      window: '24h',
    })
    expect(Array.isArray(result)).toBe(true)
  })
})
