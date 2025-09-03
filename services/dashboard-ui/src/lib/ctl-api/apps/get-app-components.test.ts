import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getAppComponents } from './get-app-components'

describe('getAppComponents should handle response status codes from GET apps/:id/components endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  test('200 status', async () => {
    const spec = await getAppComponents({ appId, orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('name')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getAppComponents({ appId, orgId }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})