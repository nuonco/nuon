import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getRunnerJobPlan } from './get-runner-job-plan'

// TODO(nnnnat): swagger has incorrect response type
describe.skip('getRunnerJobPlan should handle response status codes from GET runner-jobs/:runnerJobId/plan endpoint', () => {
  const runnerJobId = 'test-runner-job-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getRunnerJobPlan({ runnerJobId, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('status')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      getRunnerJobPlan({ runnerJobId, orgId })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
