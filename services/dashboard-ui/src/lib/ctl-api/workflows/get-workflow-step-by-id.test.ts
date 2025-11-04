import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getWorkflowStepById } from './get-workflow-step-by-id'

describe('getWorkflowStepById should handle response status codes from GET endpoint', () => {
  const orgId = 'test-org-id'
  const workflowId = 'test-workflow-id'
  const workflowStepId = 'test-workflow-step-id'

  test('200 status', async () => {
    const { data } = await getWorkflowStepById({
      orgId,
      workflowId,
      workflowStepId,
    })
    expect(data).toHaveProperty('id')
    expect(data).toHaveProperty('name')
    expect(data).toHaveProperty('workflow_id')
    expect(data).toHaveProperty('execution_type')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getWorkflowStepById({
      orgId,
      workflowId,
      workflowStepId,
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
