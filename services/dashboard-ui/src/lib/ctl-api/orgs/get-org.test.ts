import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getOrg } from './get-org'

describe('getOrg should handle response status codes from GET orgs/current endpoint', () => {
  const orgId = 'test-id'
  test('200 status', async () => {
    const spec = await getOrg({ orgId })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('name')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getOrg({ orgId }).catch((err) => expect(err).toMatchSnapshot())
  })
})