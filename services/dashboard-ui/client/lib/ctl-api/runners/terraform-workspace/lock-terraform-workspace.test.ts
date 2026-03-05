import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { lockTerraformWorkspace } from './lock-terraform-workspace'

describe('lockTerraformWorkspace should handle response status codes from POST terraform-workspaces/:workspaceId/lock endpoint', () => {
  const terraformWorkspaceId = 'test-workspace-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await lockTerraformWorkspace({ terraformWorkspaceId, orgId })
    expect(result).toBeDefined()
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      lockTerraformWorkspace({ terraformWorkspaceId, orgId })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
