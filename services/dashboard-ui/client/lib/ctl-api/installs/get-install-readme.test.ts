import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallReadme } from './get-install-readme'

describe('getInstallReadme should handle response status codes from GET installs/:installId/readme endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getInstallReadme({ installId, orgId })
    expect(result).toHaveProperty('original')
    expect(result).toHaveProperty('readme')
  }, 60000)

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getInstallReadme({ installId, orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
