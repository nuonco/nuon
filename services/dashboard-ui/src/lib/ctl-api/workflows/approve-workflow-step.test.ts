import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
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
    const body: TApproveWorkflowStepBody = {
      note: 'Approved by test',
      response_type: 'approve',
    }

    const { data, status } = await approveWorkflowStep({
      approvalId,
      body,
      orgId,
      workflowId,
      workflowStepId,
    })
    expect(data).toHaveProperty('id')
    expect(status).toBe(201)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const body: TApproveWorkflowStepBody = {
      note: 'Test note',
      response_type: 'approve',
    }

    const { error, status } = await approveWorkflowStep({
      approvalId,
      body,
      orgId,
      workflowId,
      workflowStepId,
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
