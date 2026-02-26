import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getWorkflowSteps } from './get-workflow-steps'

describe('getWorkflowSteps should handle response status codes from GET workflows/:workflowId/steps endpoint', () => {
  const orgId = 'test-id'
  const workflowId = 'test-workflow-id'

  test('200 status with all optional params', async () => {
    const { data: steps } = await getWorkflowSteps({
      orgId,
      workflowId,
      limit: 10,
      offset: 0,
    })
    steps.forEach((step) => {
      expect(step).toHaveProperty('id')
      expect(step).toHaveProperty('name')
    })
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getWorkflowSteps({ orgId, workflowId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
