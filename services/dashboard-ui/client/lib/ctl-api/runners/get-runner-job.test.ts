import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getRunnerJob } from './get-runner-job'

describe('getRunnerJob should handle response status codes from GET runner-jobs/:id endpoint', () => {
  const runnerJobId = 'test-id'
  const orgId = 'test-id'

  test('200 status', async () => {
    const result = await getRunnerJob({ runnerJobId, orgId })
    expect(result).toBeDefined()
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getRunnerJob({ runnerJobId, orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
