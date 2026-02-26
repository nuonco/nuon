import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getRunnerJob } from './get-runner-job'

describe('getRunnerJob should handle response status codes from GET runner-jobs/:id endpoint', () => {
  const runnerJobId = 'test-id'
  const orgId = 'test-id'

  test('200 status', async () => {
    const { data: spec } = await getRunnerJob({
      runnerJobId,
      orgId,
    })
    expect(spec).toBeDefined()
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getRunnerJob({ runnerJobId, orgId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
