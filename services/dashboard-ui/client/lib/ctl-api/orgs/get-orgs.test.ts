import { describe, expect, test } from 'vitest'
import { getOrgs } from './get-orgs'

describe('getOrgs should handle response status codes from GET orgs endpoint', () => {
  test('200 status', async () => {
    const result = await getOrgs()
    expect(Array.isArray(result)).toBe(true)
  })
})
