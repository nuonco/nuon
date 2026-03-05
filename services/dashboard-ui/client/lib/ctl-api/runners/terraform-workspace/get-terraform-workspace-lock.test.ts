import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getTerraformWorkspaceLock } from './get-terraform-workspace-lock'

describe('getTerraformWorkspaceLock should handle response status codes from GET terraform-workspace/:workspaceId/lock endpoint', () => {
  const workspaceId = 'test-workspace-id'
  const orgId = 'test-id'

  test('200 status', async () => {
    const result = await getTerraformWorkspaceLock({ workspaceId, orgId })
    expect(result).toBeDefined()
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      getTerraformWorkspaceLock({ workspaceId, orgId })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
