import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getWorkflow } from './get-workflow'

describe('getWorkflow should handle response status codes from GET workflows/:id endpoint', () => {
  const orgId = 'test-id'
  const workflowId = 'test-workflow-id'

  test('200 status', async () => {
    const result = await getWorkflow({ orgId, workflowId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('name')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getWorkflow({ orgId, workflowId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
