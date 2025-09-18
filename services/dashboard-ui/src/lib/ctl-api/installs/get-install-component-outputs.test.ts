import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallComponentOutputs } from './get-install-component-outputs'

describe('getInstallComponentOutputs should handle response status codes from GET installs/:installId/components/:componentId/outputs endpoint', () => {
  const componentId = 'test-id'
  const installId = 'test-id'
  const orgId = 'test-id'
  test('200 status', async () => {
    const { status } = await getInstallComponentOutputs({
      componentId,
      installId,
      orgId,
    })
    expect(status).toBe(200)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getInstallComponentOutputs({
      componentId,
      installId,
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
