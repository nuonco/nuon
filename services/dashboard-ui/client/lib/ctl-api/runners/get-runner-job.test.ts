import { describe, expect, test } from 'vitest'
import { getRunnerJob } from './get-runner-job'

describe('getRunnerJob should handle response status codes from GET runner-jobs/:id endpoint', () => {
  const runnerJobId = 'test-id'
  const orgId = 'test-id'

  test('200 status', async () => {
    const result = await getRunnerJob({ runnerJobId, orgId })
    expect(result).toBeDefined()
  })
})
