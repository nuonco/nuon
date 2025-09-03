import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { postOrg } from './post-org'

describe('postOrg should handle response status codes from POST orgs endpoint', () => {
  test('200 status', async () => {
    const spec = await postOrg({ name: 'test-name' })
    expect(spec).toHaveProperty('id')
    expect(spec).toHaveProperty('name')
  })

  test.each(badResponseCodes)('%s status', async () => {
    await postOrg({ name: 'test-name' }).catch((err) =>
      expect(err).toMatchSnapshot()
    )
  })
})