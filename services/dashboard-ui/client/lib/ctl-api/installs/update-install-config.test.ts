import { describe, expect, test } from 'vitest'
import {
  updateInstallConfig,
  type TUpdateInstallConfigBody,
} from './update-install-config'

describe('updateInstallConfig should handle response status codes from PATCH installs/:id/configs/:configId endpoint', () => {
  const orgId = 'test-org-id'
  const installId = 'test-install-id'
  const installConfigId = 'test-install-config-id'

  test('201 status with approve-all option', async () => {
    const body: TUpdateInstallConfigBody = { approval_option: 'approve-all' }
    const result = await updateInstallConfig({ body, installConfigId, installId, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('approval_option')
  })
})
