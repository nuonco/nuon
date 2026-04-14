import { describe, expect, test } from 'vitest'
import { getInstallCurrentInputs } from './get-install-current-inputs'

describe('getInstallCurrentInputs should handle response status codes from GET installs/:installId/inputs/current endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getInstallCurrentInputs({ installId, orgId })
    expect(result).toHaveProperty('values')
  })
})
