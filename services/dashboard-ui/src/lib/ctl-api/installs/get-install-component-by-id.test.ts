import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallComponentById } from './get-install-component-by-id'

describe('getInstallComponentById should handle response status codes from GET installs/:installId/components/:componentId endpoint', () => {
  const installId = 'test-install-id'
  const componentId = 'test-component-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const { data: component } = await getInstallComponentById({
      installId,
      componentId,
      orgId,
    })
    expect(component).toHaveProperty('id')
    expect(component).toHaveProperty('component_id')
  }, 60000)

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getInstallComponentById({
      installId,
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
