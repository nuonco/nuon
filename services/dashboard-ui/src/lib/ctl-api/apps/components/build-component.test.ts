import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { buildComponent } from './build-component'

describe('buildComponent should handle response status codes from POST components/:componentId/build endpoint', () => {
  const componentId = 'test-component-id'
  const orgId = 'test-org-id'

  test('200 status with default body', async () => {
    const { data: build } = await buildComponent({
      componentId,
      orgId,
    })
    expect(build).toHaveProperty('id')
    expect(build).toHaveProperty('status_v2')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await buildComponent({
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
