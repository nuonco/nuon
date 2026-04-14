import { describe, expect, test } from 'vitest'
import {
  approveAllWorkflowSteps,
  type TApproveAllWorkflowStepsBody,
} from './approve-all-workflow-steps'

describe('approveAllWorkflowSteps should handle response status codes from PATCH workflows/:id endpoint', () => {
  const orgId = 'test-org-id'
  const workflowId = 'test-workflow-id'

  test('200 status with approve-all option', async () => {
    const body: TApproveAllWorkflowStepsBody = { approval_option: 'approve-all' }
    const result = await approveAllWorkflowSteps({ body, orgId, workflowId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('status')
  })
})
