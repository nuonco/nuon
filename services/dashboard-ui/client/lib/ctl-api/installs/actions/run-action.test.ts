import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { runAction } from './run-action'

describe('runAction should handle response status codes from POST installs/:installId/action-workflows/runs endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'
  const actionWorkflowConfigId = 'test-config-id'

  test('201 status with run_env_vars', async () => {
    const { data, status } = await runAction({
      installId,
      orgId,
      body: {
        action_workflow_config_id: actionWorkflowConfigId,
        run_env_vars: {
          ENV_VAR_1: 'value1',
          ENV_VAR_2: 'value2',
        },
      },
    })
    expect(data).toEqual(expect.any(String))
    expect(status).toBe(201)
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await runAction({
      installId,
      orgId,
      body: {
        action_workflow_config_id: actionWorkflowConfigId,
        run_env_vars: {
          TEST_VAR: 'test_value',
        },
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
