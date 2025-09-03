import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getAppLatestSandboxConfig } from './get-app-latest-sandbox-config'

describe('getAppLatestSandboxConfig should handle response status codes from GET apps/:id/sandbox-latest-config endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  test('200 status', async () => {
    const spec = await getAppLatestSandboxConfig({ appId, orgId })
    expect(spec).toHaveProperty('id')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getAppLatestSandboxConfig({ appId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})