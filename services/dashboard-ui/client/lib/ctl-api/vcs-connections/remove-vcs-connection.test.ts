import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { removeVCSConnection } from './remove-vcs-connection'

describe('removeVCSConnection should handle response status codes from DELETE vcs/connections/:id endpoint', () => {
  const orgId = 'test-org-id'
  const connectionId = 'test-connection-id'

  test('204 status', async () => {
    const { status } = await removeVCSConnection({ orgId, connectionId })
    expect(status).toBe(204)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await removeVCSConnection({ orgId, connectionId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
