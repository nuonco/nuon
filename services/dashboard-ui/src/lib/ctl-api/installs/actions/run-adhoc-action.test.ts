import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { runAdhocAction } from './run-adhoc-action'

describe('runAdhocAction should handle response status codes from POST installs/:installId/actions/adhoc endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'

  test('201 status', async () => {
    const { data, status } = await runAdhocAction({
      installId,
      orgId,
      body: {
        command: 'echo "Hello, world!"',
        name: 'Test Action',
        env_vars: {
          ENV_VAR_1: 'value1',
        },
        timeout: 300,
      },
    })
    expect(data).toEqual(expect.any(String))
    expect(status).toBe(201)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await runAdhocAction({
      installId,
      orgId,
      body: {
        command: 'echo "test"',
      },
    })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
