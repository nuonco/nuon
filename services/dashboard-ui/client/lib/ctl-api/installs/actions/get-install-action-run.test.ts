import { describe, expect, test } from 'vitest'
import { getInstallActionRun } from './get-install-action-run'

describe('getInstallActionRun should handle response status codes from GET installs/:id/action-workflows/runs/:runId endpoint', () => {
  const installId = 'test-install-id'
  const runId = 'test-run-id'
  const orgId = 'test-org-id'

  test('200 status', async () => {
    const result = await getInstallActionRun({ installId, runId, orgId })
    expect(result).toHaveProperty('id')
  })
})
