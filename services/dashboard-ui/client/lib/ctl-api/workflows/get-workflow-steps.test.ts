import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getWorkflowSteps } from './get-workflow-steps'

describe('getWorkflowSteps should handle response status codes from GET workflows/:workflowId/steps endpoint', () => {
  const orgId = 'test-id'
  const workflowId = 'test-workflow-id'

  test('200 status with all optional params', async () => {
    const result = await getWorkflowSteps({ orgId, workflowId, limit: 10, offset: 0 })
    expect(Array.isArray(result)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getWorkflowSteps({ orgId, workflowId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
