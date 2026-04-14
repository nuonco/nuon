import { describe, expect, test } from 'vitest'
import { getInstall } from './get-install'

describe('getInstall should handle response status codes from GET installs/:id endpoint', () => {
  const installId = 'test-id'
  const orgId = 'test-id'

  test('200 status', async () => {
    const result = await getInstall({ installId, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('name')
  })
})
