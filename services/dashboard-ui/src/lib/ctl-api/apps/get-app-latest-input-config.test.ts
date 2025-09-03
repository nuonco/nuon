import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getAppLatestInputConfig } from './get-app-latest-input-config'

describe('getAppLatestInputConfig should handle response status codes from GET apps/:id/input-latest-config endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  test('200 status', async () => {
    const spec = await getAppLatestInputConfig({ appId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('inputs')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getAppLatestInputConfig({ appId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})