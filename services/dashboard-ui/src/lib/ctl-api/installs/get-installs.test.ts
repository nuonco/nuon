import '@test/mock-fetch-options'
import { badResponseCodes } from '@test/utils'
import { describe, expect, test } from 'vitest'
import { getInstalls } from './get-installs'

describe.skip('getInstalls should handle response status codes from GET installs endpoint', () => {
  const orgId = 'test-id'
  test('200 status', async () => {
    const spec = await getInstalls({ orgId })
    spec.forEach((s) => {
      expect(s).toHaveProperty('id')
      expect(s).toHaveProperty('name')
      expect(s).toHaveProperty('composite_component_status')
    })
  }, 60000)

  test.each(badResponseCodes)('%s status', async () => {
    await getInstalls({ orgId }).catch((err) => expect(err).toMatchSnapshot())
  })
})
