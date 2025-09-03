import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getApps } from './get-apps'

describe('getApps should handle response status codes from GET apps endpoint', () => {
  const orgId = 'test-id'
  test('200 status', async () => {
    const spec = await getApps({ orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('name')
      expect(s).toHaveProperty('cloud_platform')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getApps({ orgId }).catch((err) => expect(err).toMatchSnapshot())
  })
})