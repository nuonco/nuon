import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { createOrg, type TCreateOrgBody } from './create-org'

describe('createOrg should handle response status codes from POST orgs endpoint', () => {
  const validBody: TCreateOrgBody = {
    name: 'Test Organization',
    use_sandbox_mode: true,
  }

  test('201 status', async () => {
    const { data: org } = await createOrg({ body: validBody })
    expect(org).toHaveProperty('id')
    expect(org).toHaveProperty('name')
    expect(org).toHaveProperty('sandbox_mode')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await createOrg({ body: validBody })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
