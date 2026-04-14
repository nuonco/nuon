import { describe, expect, test } from 'vitest'
import { unlockTerraformWorkspace } from './unlock-terraform-workspace'

describe('unlockTerraformWorkspace should handle response status codes from POST terraform-workspaces/:workspaceId/unlock endpoint', () => {
  const terraformWorkspaceId = 'test-workspace-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await unlockTerraformWorkspace({ terraformWorkspaceId, orgId })
    expect(result).toBeDefined()
  })
})
