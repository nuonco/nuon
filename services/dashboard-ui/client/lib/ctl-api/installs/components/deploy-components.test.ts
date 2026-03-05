import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { deployComponents } from './deploy-components'

describe('deployComponents should handle response status codes from POST installs/:installId/components/deploy-all endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status with default body', async () => {
    const result = await deployComponents({
      body: { plan_only: true },
      installId,
      orgId,
    })
    expect(result).toBeDefined()
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      deployComponents({ body: { plan_only: true }, installId, orgId })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
