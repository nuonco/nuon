import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getOrgs } from './get-orgs'

describe('getOrgs should handle response status codes from GET orgs endpoint', () => {
  test('200 status', async () => {
    const spec = await getOrgs()
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('name')
    })
  })

  test.each(badResponseCodes)('%s status', async () => {
    await getOrgs().catch((err) => expect(err).toMatchSnapshot())
  })
})