import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { cancelRunnerJob, type ICancelRunnerJob } from './cancel-runner-job'

describe('cancelRunnerJob should handle response status codes from POST runner-jobs/:id/cancel endpoint', () => {
  const orgId = 'test-org-id'
  const runnerJobId = 'test-runner-job-id'

  test('202 status', async () => {
    const params: ICancelRunnerJob = {
      orgId,
      runnerJobId,
    }

    const { data, status } = await cancelRunnerJob(params)
    expect(data).toHaveProperty('id')
    expect(data).toHaveProperty('status')
    expect(status).toBe(202)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const params: ICancelRunnerJob = {
      orgId,
      runnerJobId,
    }

    const { error, status } = await cancelRunnerJob(params)
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
