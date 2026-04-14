import { describe, expect, test } from 'vitest'
import { runAction } from './run-action'

describe('runAction should handle response status codes from POST installs/:installId/action-workflows/runs endpoint', () => {
  const installId = 'test-install-id'
  const orgId = 'test-org-id'
  const actionWorkflowConfigId = 'test-config-id'

  test('201 status with run_env_vars', async () => {
    const result = await runAction({
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
    expect(result).toHaveProperty('data')
  })
})
