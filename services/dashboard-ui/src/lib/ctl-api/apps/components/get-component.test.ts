import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getComponent } from './get-component'

describe('getComponent should handle response status codes from GET components/:componentId endpoint', () => {
  const componentId = 'test-component-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const { data: build } = await getComponent({
      componentId,
      orgId,
    })
    expect(build).toHaveProperty('id')
    expect(build).toHaveProperty('status')
  }, 60000)

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getComponent({
      componentId,
      orgId,
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
