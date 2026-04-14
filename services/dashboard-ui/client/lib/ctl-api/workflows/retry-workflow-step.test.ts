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
    const body: TRetryWorkflowStepBody = { operation: 'skip-step', step_id: stepId }
    const result = await retryWorkflowStep({ body, orgId, workflowId })
    expect(result).toHaveProperty('workflow_id')
  })
})
