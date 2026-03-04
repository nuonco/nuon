import { badResponseCodes } from '@test/utils'
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

  test.each(badResponseCodes)('%s status', async () => {
    await expect(getInstall({ installId, orgId })).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
