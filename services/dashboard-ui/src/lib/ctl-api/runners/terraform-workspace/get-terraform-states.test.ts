import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getTerraformStates } from './get-terraform-states'

describe('getTerraformStates should handle response status codes from GET runners/terraform-workspace/:id/state-json endpoint', () => {
  const workspaceId = 'test-id'
  const orgId = 'test-id'

  test('200 status with pagination', async () => {
    const { data: spec } = await getTerraformStates({
      workspaceId,
      orgId,
      limit: 10,
      offset: 0,
    })
    expect(Array.isArray(spec)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getTerraformStates({ workspaceId, orgId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
