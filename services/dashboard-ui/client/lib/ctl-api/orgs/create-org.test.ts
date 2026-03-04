import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { createOrg, type TCreateOrgBody } from './create-org'

describe('createOrg should handle response status codes from POST orgs endpoint', () => {
  const validBody: TCreateOrgBody = {
    name: 'Test Organization',
    use_sandbox_mode: true,
  }

  test('201 status', async () => {
    const result = await createOrg({ body: validBody })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('name')
    expect(result).toHaveProperty('sandbox_mode')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(createOrg({ body: validBody })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
