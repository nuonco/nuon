import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getAppLatestConfig } from './get-app-latest-config'

describe('getAppLatestConfig should handle response status codes from GET apps/:id/latest-config endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  test('200 status', async () => {
    const spec = await getAppLatestConfig({ appId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('version')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getAppLatestConfig({ appId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})