import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getComponentBuilds } from './get-component-builds'

describe('getComponentBuilds should handle response status codes from GET /builds?:params endpoint', () => {
  const orgId = 'test-id'
  const componentId = 'test-id'
  test('200 status', async () => {
    const { data: spec } = await getComponentBuilds({
      componentId,
      orgId,
      limit: 10,
      offset: 0,
    })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('status')
    })
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getComponentBuilds({ componentId, orgId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
