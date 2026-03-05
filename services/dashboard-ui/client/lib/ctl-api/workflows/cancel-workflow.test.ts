import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { cancelWorkflow } from './cancel-workflow'

describe('cancelWorkflow should handle response status codes from POST workflows/:id/cancel endpoint', () => {
  const orgId = 'test-org-id'
  const workflowId = 'test-workflow-id'

  test('200 status', async () => {
    const result = await cancelWorkflow({ orgId, workflowId })
    expect(result).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(cancelWorkflow({ orgId, workflowId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
