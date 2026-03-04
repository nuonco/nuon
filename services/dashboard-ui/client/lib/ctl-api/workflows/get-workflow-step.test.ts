import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getWorkflowStep } from './get-workflow-step'

describe('getWorkflowStep should handle response status codes from GET endpoint', () => {
  const orgId = 'test-org-id'
  const workflowId = 'test-workflow-id'
  const workflowStepId = 'test-workflow-step-id'

  test('200 status', async () => {
    const result = await getWorkflowStep({ orgId, workflowId, workflowStepId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('name')
    expect(result).toHaveProperty('workflow_id')
    expect(result).toHaveProperty('execution_type')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      getWorkflowStep({ orgId, workflowId, workflowStepId })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
