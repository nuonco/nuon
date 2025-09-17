import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getRunnerJobPlanById } from './get-runner-job-plan-by-id'

// TODO(nnnnat): swagger has incorrect response type
describe.skip('getRunnerJobPlanById should handle response status codes from GET runner-jobs/:runnerJobId/plan endpoint', () => {
  const runnerJobId = 'test-runner-job-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const { data: plan } = await getRunnerJobPlanById({
      runnerJobId,
      orgId,
    })
    expect(plan).toHaveProperty('id')
    expect(plan).toHaveProperty('status')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getRunnerJobPlanById({
      runnerJobId,
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
