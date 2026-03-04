import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { runAdhocAction } from './run-adhoc-action'

describe('runAdhocAction should handle response status codes from POST installs/:installId/actions/adhoc endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('201 status', async () => {
    const result = await runAdhocAction({
      installId,
      orgId,
      body: {
        command: 'echo "Hello, world!"',
        name: 'Test Action',
        env_vars: { ENV_VAR_1: 'value1' },
        timeout: 300,
      },
    })
    expect(result).toHaveProperty('data')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await expect(
      runAdhocAction({ installId, orgId, body: { command: 'echo "test"' } })
    ).rejects.toMatchObject({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
