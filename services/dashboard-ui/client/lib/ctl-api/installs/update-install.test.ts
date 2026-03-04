import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { updateInstall } from './update-install'

describe('updateInstall should handle response status codes from PATCH installs/:installId endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status with metadata managed_by dashboard', async () => {
    const result = await updateInstall({
      installId,
      orgId,
      body: { metadata: { managed_by: 'nuon/dashboard' } },
    })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('name')
    expect(result).toHaveProperty('app_id')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      updateInstall({
        installId,
        orgId,
        body: { install_config: { approval_option: 'prompt' } },
      })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
