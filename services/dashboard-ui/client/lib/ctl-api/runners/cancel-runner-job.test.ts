import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { cancelRunnerJob } from './cancel-runner-job'

describe('cancelRunnerJob should handle response status codes from POST runner-jobs/:id/cancel endpoint', () => {
  const orgId = 'test-org-id'
  const runnerJobId = 'test-runner-job-id'

  test('202 status', async () => {
    const result = await cancelRunnerJob({ orgId, runnerJobId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('status')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(cancelRunnerJob({ orgId, runnerJobId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
