import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallCurrentInputs } from './get-install-current-inputs'

describe('getInstallCurrentInputs should handle response status codes from GET installs/:installId/inputs/current endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const { data: run, status } = await getInstallCurrentInputs({
      installId,
      orgId,
    })
    expect(status).toBe(200)
    expect(run).toHaveProperty('values')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getInstallCurrentInputs({
      installId,
      orgId,
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
