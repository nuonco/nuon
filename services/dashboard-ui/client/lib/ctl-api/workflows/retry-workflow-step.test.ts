import { describe, expect, test } from 'vitest'
import { retryWorkflowStep } from './retry-workflow-step'

describe('retryWorkflowStep', () => {
  const orgId = 'test-org-id'
  const workflowId = 'test-workflow-id'
  const stepId = 'test-step-id'

  test('calls the correct endpoint', async () => {
    const result = await retryWorkflowStep({ orgId, workflowId, stepId })
    expect(result).toHaveProperty('workflow_id')
  })
})
