import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getAppLatestRunnerConfig } from './get-app-latest-runner-config'

describe('getAppLatestRunnerConfig should handle response status codes from GET apps/:id/runner-latest-config endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  test('200 status', async () => {
    const spec = await getAppLatestRunnerConfig({ appId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('app_runner_type')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getAppLatestRunnerConfig({ appId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})