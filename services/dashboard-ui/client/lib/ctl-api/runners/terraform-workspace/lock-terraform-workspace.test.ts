import { describe, expect, test } from 'vitest'
import { lockTerraformWorkspace } from './lock-terraform-workspace'

describe('lockTerraformWorkspace should handle response status codes from POST terraform-workspaces/:workspaceId/lock endpoint', () => {
  const terraformWorkspaceId = 'test-workspace-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await lockTerraformWorkspace({ terraformWorkspaceId, orgId })
    expect(result).toBeDefined()
  })
})
