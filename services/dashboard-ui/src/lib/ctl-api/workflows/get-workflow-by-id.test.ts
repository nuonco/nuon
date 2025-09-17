import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getWorkflowById } from './get-workflow-by-id'

describe('getWorkflowById should handle response status codes from GET workflows/:id endpoint', () => {
  const orgId = 'test-id'
  const workflowId = 'test-workflow-id'

  test('200 status', async () => {
    const { data: workflow } = await getWorkflowById({ orgId, workflowId })
    expect(workflow).toHaveProperty('id')
    expect(workflow).toHaveProperty('name')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getWorkflowById({ orgId, workflowId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
