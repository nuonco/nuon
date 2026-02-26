import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import {
  retryWorkflowStep,
  type TRetryWorkflowStepBody,
} from './retry-workflow-step'

describe('retryWorkflowStep should handle response status codes from POST workflows/:id/retry endpoint', () => {
  const orgId = 'test-org-id'
  const workflowId = 'test-workflow-id'
  const stepId = 'test-step-id'

  test('201 status with skip-step operation', async () => {
    const body: TRetryWorkflowStepBody = {
      operation: 'skip-step',
      step_id: stepId,
    }

    const { data, status } = await retryWorkflowStep({
      body,
      orgId,
      workflowId,
    })
    expect(data).toHaveProperty('workflow_id')
    expect(status).toBe(201)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const body: TRetryWorkflowStepBody = {
      operation: 'retry-step',
      step_id: stepId,
    }

    const { error, status } = await retryWorkflowStep({
      body,
      orgId,
      workflowId,
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
