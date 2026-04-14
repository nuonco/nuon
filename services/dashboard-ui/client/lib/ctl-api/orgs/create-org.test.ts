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
})
