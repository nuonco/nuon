import { describe, expect, test } from 'vitest'
import {
  createInstallConfig,
  type TCreateInstallConfigBody,
} from './create-install-config'

describe('createInstallConfig should handle response status codes from POST installs/:id/config endpoint', () => {
  const orgId = 'test-org-id'
  const installId = 'test-install-id'

  test('201 status with approve-all option', async () => {
    const body: TCreateInstallConfigBody = { approval_option: 'approve-all' }
    const result = await createInstallConfig({ body, installId, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('approval_option')
  })
})
