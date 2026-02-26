import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstalls } from './get-installs'

describe('getInstalls should handle response status codes from GET installs endpoint', () => {
  const orgId = 'test-id'

  test('200 status with pagination params', async () => {
    const { data: installs } = await getInstalls({
      orgId,
      limit: 10,
      offset: 0,
    })
    expect(Array.isArray(installs)).toBe(true)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getInstalls({ orgId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
