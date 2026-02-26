import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { cancelWorkflow } from './cancel-workflow'

describe('cancelWorkflow should handle response status codes from POST workflows/:id/cancel endpoint', () => {
  const orgId = 'test-org-id'
  const workflowId = 'test-workflow-id'

  test('200 status', async () => {
    const { data } = await cancelWorkflow({ orgId, workflowId })
    expect(data).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await cancelWorkflow({ orgId, workflowId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
