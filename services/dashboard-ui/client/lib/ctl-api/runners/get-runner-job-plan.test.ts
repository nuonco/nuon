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
})
