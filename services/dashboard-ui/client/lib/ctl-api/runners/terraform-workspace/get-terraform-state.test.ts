import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getTerraformState } from './get-terraform-state'

describe('getTerraformState should handle response status codes from GET runners/terraform-workspace/:workspaceId/state-json/:stateId endpoint', () => {
  const workspaceId = 'test-workspace-id'
  const stateId = 'test-state-id'
  const orgId = 'test-id'

  test('200 status', async () => {
    const { data: spec } = await getTerraformState({
      workspaceId,
      stateId,
      orgId,
    })
    expect(spec).toBeDefined()
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getTerraformState({
      workspaceId,
      stateId,
      orgId,
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
