import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getComponentDeploys } from './get-component-deploys'

describe('getComponentDeploys should handle response status codes from GET installs/:installId/components/:componentId/deploys endpoint', () => {
  const installId = 'test-install-id'
  const componentId = 'test-component-id'
  const orgId = 'test-org-id'

  test('200 status with pagination', async () => {
    const result = await getComponentDeploys({
      installId,
      componentId,
      orgId,
      limit: 10,
      offset: 0,
    })
    expect(Array.isArray(result)).toBe(true)
  }, 60000)

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      getComponentDeploys({ installId, componentId, orgId })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
