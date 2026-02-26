import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstallReadme } from './get-install-readme'

describe('getInstallReadme should handle response status codes from GET installs/:installId/readme endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const { data: readme } = await getInstallReadme({
      installId,
      orgId,
    })
    expect(readme).toHaveProperty('original')
    expect(readme).toHaveProperty('readme')
  }, 60000)

  test('206 status', async () => {
    const { data: readme } = await getInstallReadme({
      installId,
      orgId,
    })
    expect(readme).toHaveProperty('original')
    expect(readme).toHaveProperty('readme')
    expect(readme).toHaveProperty('warnings')
  }, 60000)

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getInstallReadme({
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
