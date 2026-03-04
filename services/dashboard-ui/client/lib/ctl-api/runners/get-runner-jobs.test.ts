import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getRunnerJobs } from './get-runner-jobs'

describe('getRunnerJobs should handle response status codes from GET runners/:id/jobs endpoint', () => {
  const runnerId = 'test-id'
  const orgId = 'test-id'

  test('200 status with pagination', async () => {
    const result = await getRunnerJobs({ runnerId, orgId, limit: 10, offset: 0 })
    expect(Array.isArray(result)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getRunnerJobs({ runnerId, orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
