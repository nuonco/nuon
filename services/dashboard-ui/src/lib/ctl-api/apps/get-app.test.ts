import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getApp } from './get-app'

describe('getApp should handle response status codes from GET apps/:id endpoint', () => {
  const orgId = 'test-id'
  const appId = 'test-id'
  test('200 status', async () => {
    const spec = await getApp({ appId, orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('name')
    expect(spec).toHaveProperty('cloud_platform')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getApp({ appId, orgId }).catch((err) => expect(err).toMatchSnapshot())
  })
})