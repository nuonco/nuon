import '@test/mock-auth'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getRunnerSettingsById } from './get-runner-settings-by-id'

describe('getRunnerSettingsById should handle response status codes from GET runners/:id/settings endpoint', () => {
  const runnerId = 'test-id'
  const orgId = 'test-id'
  test('200 status', async () => {
    const { data: runner } = await getRunnerSettingsById({ runnerId, orgId })
    expect(runner).toHaveProperty('id')
    expect(runner).toHaveProperty('created_at')
  })

  test.each(badResponseCodes)('%s status', async (code) => {
    const { error, status } = await getRunnerSettingsById({ runnerId, orgId })
    expect(status).toBe(code)
    expect(error).toMatchSnapshot({
      error: expect.any(String),
      description: expect.any(String),
      user_error: expect.any(Boolean),
    })
  })
})
