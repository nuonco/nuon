import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { nueGetOrg } from './nue-get-org'

describe('nueGetOrg should handle response status codes from GET orgs/current endpoint', () => {
  const orgId = 'test-id'
  test('200 status', async () => {
    const { data, error, status } = await nueGetOrg({ orgId })
    expect(data).toHaveProperty("id")
    expect(data).toHaveProperty("name")
    expect(error).toBeNull()
    expect(status).toBe(200)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { data, error, status } = await nueGetOrg({ orgId })
    expect(data).toBeNull()
    expect(error).toHaveProperty("description")
    expect(error).toHaveProperty("error")
    expect(error).toHaveProperty("user_error")
    expect(status).toBe(code)
  })
})