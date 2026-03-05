import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { removeVCSConnection } from './remove-vcs-connection'

describe('removeVCSConnection should handle response status codes from DELETE vcs/connections/:id endpoint', () => {
  const orgId = 'test-org-id'
  const connectionId = 'test-connection-id'

  test('204 status', async () => {
    const result = await removeVCSConnection({ orgId, connectionId })
    expect(result).toBeNull()
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      removeVCSConnection({ orgId, connectionId })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
