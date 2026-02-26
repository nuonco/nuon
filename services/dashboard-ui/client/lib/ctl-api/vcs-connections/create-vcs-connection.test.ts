import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { createVCSConnection } from './create-vcs-connection'

describe('createVCSConnection should handle response status codes from POST vcs/connections endpoint', () => {
  const orgId = 'test-org-id'
  const body = { github_install_id: 'test-github-install-id' }

  test('201 status', async () => {
    const { data, status } = await createVCSConnection({ orgId, body })
    expect(status).toBe(201)
    expect(data).toHaveProperty('id')
    expect(data).toHaveProperty('github_install_id')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await createVCSConnection({ orgId, body })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
