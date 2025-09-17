import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getRunnerJobById } from './get-runner-job-by-id'

describe('getRunnerJobById should handle response status codes from GET runner-jobs/:id endpoint', () => {
  const runnerJobId = 'test-id'
  const orgId = 'test-id'

  test('200 status', async () => {
    const { data: spec } = await getRunnerJobById({
      runnerJobId,
      orgId,
    })
    expect(spec).toBeDefined()
    expect(spec).toMatchSnapshot()
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getRunnerJobById({ runnerJobId, orgId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
