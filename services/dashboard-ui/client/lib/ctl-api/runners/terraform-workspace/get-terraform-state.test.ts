import { describe, expect, test } from 'vitest'
import { getTerraformState } from './get-terraform-state'

describe('getTerraformState should handle response status codes from GET runners/terraform-workspace/:workspaceId/state-json/:stateId endpoint', () => {
  const workspaceId = 'test-workspace-id'
  const stateId = 'test-state-id'
  const orgId = 'test-id'

  test('200 status', async () => {
    const result = await getTerraformState({ workspaceId, stateId, orgId })
    expect(result).toBeDefined()
  })
})
