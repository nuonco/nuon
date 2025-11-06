import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallState } from './get-install-state'

describe('getInstallState should handle response status codes from GET installs/:id/state endpoint', () => {
  const installId = 'test-id'
  const orgId = 'test-id'
  test('200 status', async () => {
    const { data: state } = await getInstallState({ installId, orgId })
    expect(state).toBeDefined()
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getInstallState({ installId, orgId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
