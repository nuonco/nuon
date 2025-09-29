import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallStack } from './get-install-stack'

describe('getInstallStack should handle response status codes from GET installs/:id/stack endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'
  
  test('200 status', async () => {
    const { data: stack } = await getInstallStack({ installId, orgId })
    expect(stack).toBeDefined()
    expect(stack).toHaveProperty('id')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getInstallStack({ installId, orgId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})