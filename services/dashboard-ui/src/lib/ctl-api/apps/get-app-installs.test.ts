import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getAppInstalls } from './get-app-installs'

describe('getAppInstalls should handle response status codes from GET apps/:id/installs endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  test('200 status', async () => {
    const spec = await getAppInstalls({ appId, orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('name')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getAppInstalls({ appId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})