import { describe, expect, test } from 'vitest'
import { getRunnerJobs } from './get-runner-jobs'

describe('getRunnerJobs should handle response status codes from GET runners/:id/jobs endpoint', () => {
  const runnerId = 'test-id'
  const orgId = 'test-id'

  test('200 status with pagination', async () => {
    const result = await getRunnerJobs({ runnerId, orgId, limit: 10, offset: 0 })
    expect(Array.isArray(result.data)).toBe(true)
  })
})
