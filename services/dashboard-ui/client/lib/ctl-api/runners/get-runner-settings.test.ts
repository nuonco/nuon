import { describe, expect, test } from 'vitest'
import { getRunnerSettings } from './get-runner-settings'

describe('getRunnerSettings should handle response status codes from GET runners/:id/settings endpoint', () => {
  const runnerId = 'test-id'
  const orgId = 'test-id'

  test('200 status', async () => {
    const result = await getRunnerSettings({ runnerId, orgId })
    expect(result).toHaveProperty('id')
    expect(result).toHaveProperty('created_at')
  })
})
