import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { updateInstall } from './update-install'

describe('updateInstall should handle response status codes from PATCH installs/:installId endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status with metadata managed_by dashboard', async () => {
    const { data: install } = await updateInstall({
      installId,
      orgId,
      body: {
        metadata: {
          managed_by: 'nuon/dashboard',
        },
      },
    })
    expect(install).toHaveProperty('id')
    expect(install).toHaveProperty('name')
    expect(install).toHaveProperty('app_id')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await updateInstall({
      installId,
      orgId,
      body: {
        install_config: {
          approval_option: 'prompt',
        },
      },
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
