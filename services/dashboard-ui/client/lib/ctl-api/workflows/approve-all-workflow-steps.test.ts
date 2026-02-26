import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import {
  approveAllWorkflowSteps,
  type TApproveAllWorkflowStepsBody,
} from './approve-all-workflow-steps'

describe('approveAllWorkflowSteps should handle response status codes from PATCH workflows/:id endpoint', () => {
  const orgId = 'test-org-id'
  const workflowId = 'test-workflow-id'

  test('200 status with approve-all option', async () => {
    const body: TApproveAllWorkflowStepsBody = {
      approval_option: 'approve-all',
    }

    const { data } = await approveAllWorkflowSteps({
      body,
      orgId,
      workflowId,
    })
    expect(data).toHaveProperty('id')
    expect(data).toHaveProperty('status')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const body: TApproveAllWorkflowStepsBody = {
      approval_option: 'approve-all',
    }

    const { error, status } = await approveAllWorkflowSteps({
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
