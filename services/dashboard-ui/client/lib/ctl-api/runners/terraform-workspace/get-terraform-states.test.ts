import { describe, expect, test } from 'vitest'
import { getTerraformStates } from './get-terraform-states'

describe('getTerraformStates should handle response status codes from GET runners/terraform-workspace/:id/state-json endpoint', () => {
  const workspaceId = 'test-id'
  const orgId = 'test-id'

  test('200 status with pagination', async () => {
    const result = await getTerraformStates({ workspaceId, orgId, limit: 10, offset: 0 })
    expect(Array.isArray(result)).toBe(true)
  })
})
