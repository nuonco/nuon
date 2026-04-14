import { describe, expect, test } from 'vitest'
import {
  approveWorkflowStep,
  type TApproveWorkflowStepBody,
} from './approve-workflow-step'

describe('approveWorkflowStep should handle response status codes from POST workflows/:workflowId/steps/:workflowStepId/approvals/:approvalId/response endpoint', () => {
  const approvalId = 'test-approval-id'
  const orgId = 'test-org-id'
  const workflowId = 'test-workflow-id'
  const workflowStepId = 'test-workflow-step-id'

  test('201 status with approve operation', async () => {
    const body: TApproveWorkflowStepBody = { note: 'Approved by test', response_type: 'approve' }
    const result = await approveWorkflowStep({ approvalId, body, orgId, workflowId, workflowStepId })
    expect(result).toHaveProperty('id')
  })
})
