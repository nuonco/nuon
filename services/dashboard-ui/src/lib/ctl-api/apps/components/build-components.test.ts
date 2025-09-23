import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { buildComponents } from './build-components'

// TODO(nnnnat): need to figure out a good way to test this
describe.skip('buildComponents should handle response status codes from POST components/:componentId/build endpoint', () => {
  const components = [{ id: 'test-component-id' }, { id: 'test-component-id' }]
  const orgId = 'test-org-id'

  test('200 status with default body', async () => {
    const [{ data: build }] = await buildComponents({
      components,
      orgId,
    })
    expect(build).toHaveProperty('id')
    expect(build).toHaveProperty('status_v2')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const [{ error, status }] = await buildComponents({
      components,
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
