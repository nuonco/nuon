import { describe, expect, test } from 'vitest'
import { getInstallSandboxRun } from './get-install-sandbox-run'

describe('getInstallSandboxRun should handle response status codes from GET installs/sandbox-runs/:runId endpoint', () => {
  const runId = 'test-run-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getInstallSandboxRun({ runId, orgId })
    expect(result).toHaveProperty('id')
  })
})
