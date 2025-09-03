import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getAppConfigs } from './get-app-configs'

describe('getAppConfigs should handle response status codes from GET apps/:id/configs endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  test('200 status', async () => {
    const spec = await getAppConfigs({ appId, orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('version')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getAppConfigs({ appId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})