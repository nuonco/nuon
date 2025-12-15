import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getComponentBuild } from './get-component-build'

describe('getComponentBuild should handle response status codes from GET components/:componentId/builds/:buildId endpoint', () => {
  const componentId = 'test-component-id'
  const buildId = 'test-build-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const { data: build } = await getComponentBuild({
      componentId,
      buildId,
      orgId,
    })
    expect(build).toHaveProperty('id')
    expect(build).toHaveProperty('status')
  }, 60000)

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getComponentBuild({
      componentId,
      buildId,
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
