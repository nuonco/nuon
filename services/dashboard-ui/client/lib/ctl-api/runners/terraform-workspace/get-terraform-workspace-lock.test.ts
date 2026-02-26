import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getTerraformWorkspaceLock } from './get-terraform-workspace-lock'

describe('getTerraformWorkspaceLock should handle response status codes from GET terraform-workspace/:workspaceId/lock endpoint', () => {
  const workspaceId = 'test-workspace-id'
  const orgId = 'test-id'

  test('200 status', async () => {
    const { data: spec } = await getTerraformWorkspaceLock({
      workspaceId,
      orgId,
    })
    expect(spec).toBeDefined()
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getTerraformWorkspaceLock({
      workspaceId,
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
