import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getOrgs } from './get-orgs'

describe('getOrgs should handle response status codes from GET orgs endpoint', () => {
  test('200 status', async () => {
    const { data: spec } = await getOrgs()
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('name')
    })
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getOrgs()
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
