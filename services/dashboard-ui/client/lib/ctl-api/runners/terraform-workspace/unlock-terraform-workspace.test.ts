import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { unlockTerraformWorkspace } from './unlock-terraform-workspace'

describe('unlockTerraformWorkspace should handle response status codes from POST terraform-workspaces/:workspaceId/unlock endpoint', () => {
  const terraformWorkspaceId = 'test-workspace-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await unlockTerraformWorkspace({ terraformWorkspaceId, orgId })
    expect(result).toBeDefined()
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      unlockTerraformWorkspace({ terraformWorkspaceId, orgId })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
