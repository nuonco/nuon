import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstall } from './get-install'

describe('getInstall should handle response status codes from GET installs/:id endpoint', () => {
  const installId = 'test-id'
  const orgId = 'test-id'
  test('200 status', async () => {
    const { data: spec } = await getInstall({ installId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('name')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getInstall({ installId, orgId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
